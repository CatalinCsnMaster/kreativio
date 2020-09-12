// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

// Package builder provides dynamic query building based on conditions.
package builder

import (
	"fmt"
	"strings"

	"github.com/moapis/shop"
	"github.com/moapis/shop/models"
)

const (
	listQuery = `with %s
select json_agg(
	json_build_object(
		%s
	)
)
from arts a
%s;`

	mainCTE = `filters as (
	select m.id
	from %s.%s m
	%s
),
arts as (
	select %s
	from filters f
	join shop.articles a on a.id = f.id
	%s
)`

	relationCTE = `%s as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			%s
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	%s
	group by arts.id
)`

	relationJoin = `left join %s.%s %s on %s.%s = %s.%s`
)

// Debug enables printing of generated queries
var Debug bool

func leftJoin(schema, table, alias, column, la, lc string) string {
	return fmt.Sprintf(relationJoin, schema, table, alias, alias, column, la, lc)
}

type relation struct {
	name    string
	columns []string
	id      string // Relation's id column to join against

	joinTable string    // Intermediate join table to use
	joinIDs   [2]string // Left and right join columns
}

func (r relation) joins(schema string) string {
	if r.joinTable == "" {
		return leftJoin(schema, r.name, "r", r.id, "arts", "id")
	}

	return strings.Join([]string{
		leftJoin(schema, r.joinTable, "j", r.joinIDs[0], "arts", "id"),
		leftJoin(schema, r.name, "r", r.id, "j", r.joinIDs[1]),
	}, "\n\t")
}

func jboCast(alias, col string) string {
	switch col {
	case "price", "multiplier":
		return fmt.Sprintf("%s.%s::text", alias, col)
	default:
		return fmt.Sprintf("%s.%s", alias, col)
	}
}

func (r relation) build(schema, alias string) (string, []string) {
	jbo := make([]string, 0, len(r.columns)*2)
	for _, c := range r.columns {
		jbo = append(jbo, fmt.Sprintf("'%s'", c), jboCast("r", c))
	}

	return fmt.Sprintf(
			relationCTE,
			alias,
			strings.Join(jbo, ", "),
			r.joins(schema),
		),
		[]string{
			fmt.Sprintf("'%s'", r.name),
			fmt.Sprintf("%s.js", alias),
		}
}

type list struct {
	name      string
	columns   []string
	relations []relation
	limit     int32
	offset    int32
}

// Query builds and returns the Article List query
func (l *list) query(schema, filters string) string {
	var (
		sel = []string{"a.id"} // id is always required
		jbo []string
	)
	for _, col := range l.columns {
		if col != "id" {
			sel = append(sel, fmt.Sprintf("a.%s", col))
		}
		jbo = append(jbo, fmt.Sprintf("'%s'", col), jboCast("a", col))
	}

	var limits string
	if l.limit != 0 {
		limits = fmt.Sprintf("limit %d\n\toffset %d", l.limit, l.offset)
	}

	ctes := []string{fmt.Sprintf(
		mainCTE,
		schema,
		l.name,
		filters,
		strings.Join(sel, ", "),
		limits,
	)}

	joins := make([]string, len(l.relations))
	for i, r := range l.relations {
		alias := fmt.Sprintf("r%d", i)
		cte, jb := r.build(schema, alias)
		ctes = append(ctes, cte)
		jbo = append(jbo, jb...)
		joins[i] = fmt.Sprintf("join %s on a.id = %s.id", alias, alias)
	}

	return fmt.Sprintf(
		listQuery,
		strings.Join(ctes, ",\n"),
		strings.Join(jbo, ", "),
		strings.Join(joins, "\n"),
	)
}

func defaults(cond *shop.ListConditions) ([]shop.ArticleFields, *shop.ArticleRelations, *shop.Limits) {
	fields, relations := cond.GetFields(), cond.GetRelations()
	if len(fields) == 0 {
		fields = []shop.ArticleFields{
			shop.ArticleFields_ID,
			shop.ArticleFields_TITLE,
			shop.ArticleFields_PRICE,
			shop.ArticleFields_PROMOTED,
		}
	}

	if relations == nil {
		relations = &shop.ArticleRelations{
			Images: []shop.MediaFields{
				shop.MediaFields_MD_URL,
				shop.MediaFields_MD_LABEL,
			},
		}
	}

	limits := cond.GetLimits()
	switch {
	case limits == nil:
		limits = &shop.Limits{Limit: DefaultLimit}
	case limits.GetLimit() == 0:
		limits.Limit = DefaultLimit
	}

	return fields, relations, limits
}

const (
	joinArticlesCategory = "join %s.category_articles ac on ac.article_id = m.id"
	joinCategories       = "join %s.categories c on c.id = ac.category_id"
)

func filters(cond *shop.ListConditions, schema string) (filters string, args []interface{}) {
	var (
		joins, wheres []string
	)
	if cond.GetOnlyPublished() {
		wheres = append(wheres, "m.published")
	}
	if cond.GetOnlyPromoted() {
		wheres = append(wheres, "m.promoted")
	}

	if cat := cond.GetOnlyCategoryId(); cat != 0 {
		args = append(args, cat)
		joins = append(joins, fmt.Sprintf(joinArticlesCategory, schema))
		wheres = append(wheres, fmt.Sprintf("ac.category_id = $%d", len(args)))
	} else if cal := cond.GetOnlyCategoryLabel(); cal != "" {
		args = append(args, cal)
		joins = append(joins, fmt.Sprintf(joinArticlesCategory, schema))
		joins = append(joins, fmt.Sprintf(joinCategories, schema))
		wheres = append(wheres, fmt.Sprintf("c.label = $%d", len(args)))
	}

	if len(wheres) == 0 {
		return "", nil
	}

	return strings.Join([]string{
		strings.Join(joins, "\n\t"),
		fmt.Sprintf("where %s", strings.Join(wheres, "\n\tand ")),
	}, "\n\t"), args
}

// DefaultLimit is used when ListConditions does not specify any.
// Zero unsets the default.
var DefaultLimit = int32(25)

// ArticleListQuery builds and returns an SQL query based on the provided conditions.
// The query is designed to return a single JSON array as result.
// This array should be unmarshalled into a []shop.Article slice.
func ArticleListQuery(cond *shop.ListConditions, schema string) (string, []interface{}, error) {
	fields, relations, limits := defaults(cond)

	cols, err := ArticleColumns(fields)
	if err != nil {
		return "", nil, err
	}

	rels, err := articleRelations(relations)
	if err != nil {
		return "", nil, err
	}

	lq := list{
		name:      models.TableNames.Articles,
		columns:   cols,
		relations: rels,
		limit:     limits.GetLimit(),
		offset:    limits.GetOffset(),
	}

	f, args := filters(cond, schema)
	query := lq.query(schema, f)
	if Debug {
		fmt.Println(query)
	}
	return query, args, nil
}
