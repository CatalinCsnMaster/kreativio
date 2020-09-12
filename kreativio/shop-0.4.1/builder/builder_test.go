// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

package builder

import (
	"reflect"
	"testing"

	"github.com/moapis/shop"
)

const (
	imagesCTEOut = `r1 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'url', r.url, 'label', r.label
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.images r on r.article_id = arts.id
	group by arts.id
)`

	categoriesCTEOut = `r1 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'id', r.id, 'label', r.label
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.category_articles j on j.article_id = arts.id
	left join shop.categories r on r.id = j.category_id
	group by arts.id
)`
)

func init() {
	Debug = true
}

func TestRelation_build(t *testing.T) {
	type fields struct {
		name      string
		columns   []string
		id        string
		joinTable string
		joinIDs   [2]string
	}
	type args struct {
		schema string
		alias  string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  []string
	}{
		{
			"Images",
			fields{
				name:    "images",
				columns: []string{"url", "label"},
				id:      "article_id",
			},
			args{
				"shop",
				"r1",
			},
			imagesCTEOut,
			[]string{"'images'", "r1.js"},
		},
		{
			"Categories",
			fields{
				name:      "categories",
				columns:   []string{"id", "label"},
				id:        "id",
				joinTable: "category_articles",
				joinIDs:   [2]string{"article_id", "category_id"},
			},
			args{
				"shop",
				"r1",
			},
			categoriesCTEOut,
			[]string{"'categories'", "r1.js"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := relation{
				name:      tt.fields.name,
				columns:   tt.fields.columns,
				id:        tt.fields.id,
				joinTable: tt.fields.joinTable,
				joinIDs:   tt.fields.joinIDs,
			}
			got, got1 := r.build(tt.args.schema, tt.args.alias)
			if got != tt.want {
				t.Errorf("Relation.build() got = \n%v\nwant\n%v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Relation.build() got1 = \n%v\nwant\n%v", got1, tt.want1)
			}
		})
	}
}

const (
	allArtsNoLimits = `with filters as (
	select m.id
	from shop.articles m
	
),
arts as (
	select a.id, a.title, a.price
	from filters f
	join shop.articles a on a.id = f.id
	
)
select json_agg(
	json_build_object(
		'id', a.id, 'title', a.title, 'price', a.price::text
	)
)
from arts a
;`

	allArtsWithImages = `with filters as (
	select m.id
	from shop.articles m
	
),
arts as (
	select a.id, a.title, a.price
	from filters f
	join shop.articles a on a.id = f.id
	limit 25
	offset 0
),
r0 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'url', r.url, 'label', r.label
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.images r on r.article_id = arts.id
	group by arts.id
)
select json_agg(
	json_build_object(
		'id', a.id, 'title', a.title, 'price', a.price::text, 'images', r0.js
	)
)
from arts a
join r0 on a.id = r0.id;`

	allArtsWithImagesAndCategories = `with filters as (
	select m.id
	from shop.articles m
	
),
arts as (
	select a.id, a.title, a.price
	from filters f
	join shop.articles a on a.id = f.id
	limit 25
	offset 0
),
r0 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'url', r.url, 'label', r.label
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.images r on r.article_id = arts.id
	group by arts.id
),
r1 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'id', r.id, 'label', r.label
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.category_articles j on j.article_id = arts.id
	left join shop.categories r on r.id = j.category_id
	group by arts.id
)
select json_agg(
	json_build_object(
		'id', a.id, 'title', a.title, 'price', a.price::text, 'images', r0.js, 'categories', r1.js
	)
)
from arts a
join r0 on a.id = r0.id
join r1 on a.id = r1.id;`

	allArtsWithVariantsAndOffset = `with filters as (
	select m.id
	from shop.articles m
	
),
arts as (
	select a.id, a.title, a.price
	from filters f
	join shop.articles a on a.id = f.id
	limit 100
	offset 200
),
r0 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'labels', r.labels, 'multiplier', r.multiplier::text
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.variants r on r.article_id = arts.id
	group by arts.id
)
select json_agg(
	json_build_object(
		'id', a.id, 'title', a.title, 'price', a.price::text, 'variants', r0.js
	)
)
from arts a
join r0 on a.id = r0.id;`
)

func TestList_query(t *testing.T) {
	type fields struct {
		name      string
		columns   []string
		relations []relation
		limit     int32
		offset    int32
	}
	tests := []struct {
		name    string
		fields  fields
		filters string
		want    string
	}{
		{
			"All arts no limits",
			fields{
				"articles",
				[]string{"id", "title", "price"},
				nil,
				0,
				0,
			},
			"",
			allArtsNoLimits,
		},
		{
			"All arts with images",
			fields{
				"articles",
				[]string{"id", "title", "price"},
				[]relation{{
					name:    "images",
					columns: []string{"url", "label"},
					id:      "article_id",
				}},
				25,
				0,
			},
			"",
			allArtsWithImages,
		},
		{
			"All arts with images and categories",
			fields{
				"articles",
				[]string{"id", "title", "price"},
				[]relation{
					{
						name:    "images",
						columns: []string{"url", "label"},
						id:      "article_id",
					},
					{
						name:      "categories",
						columns:   []string{"id", "label"},
						id:        "id",
						joinTable: "category_articles",
						joinIDs:   [2]string{"article_id", "category_id"},
					},
				},
				25,
				0,
			},
			"",
			allArtsWithImagesAndCategories,
		},
		{
			"All arts with variants",
			fields{
				"articles",
				[]string{"id", "title", "price"},
				[]relation{{
					name:    "variants",
					columns: []string{"labels", "multiplier"},
					id:      "article_id",
				}},
				100,
				200,
			},
			"",
			allArtsWithVariantsAndOffset,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &list{
				name:      tt.fields.name,
				columns:   tt.fields.columns,
				relations: tt.fields.relations,
				limit:     tt.fields.limit,
				offset:    tt.fields.offset,
			}
			if got := l.query("shop", tt.filters); got != tt.want {
				t.Errorf("List.Query() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_defaults(t *testing.T) {
	tests := []struct {
		name  string
		cond  *shop.ListConditions
		want  []shop.ArticleFields
		want1 *shop.ArticleRelations
		want2 *shop.Limits
	}{
		{
			"Defaults",
			&shop.ListConditions{},
			[]shop.ArticleFields{
				shop.ArticleFields_ID,
				shop.ArticleFields_TITLE,
				shop.ArticleFields_PRICE,
				shop.ArticleFields_PROMOTED,
			},
			&shop.ArticleRelations{
				Images: []shop.MediaFields{
					shop.MediaFields_MD_URL,
					shop.MediaFields_MD_LABEL,
				},
			},
			&shop.Limits{
				Limit: DefaultLimit,
			},
		},
		{
			"Defaults zero limit",
			&shop.ListConditions{
				Limits: &shop.Limits{},
			},
			[]shop.ArticleFields{
				shop.ArticleFields_ID,
				shop.ArticleFields_TITLE,
				shop.ArticleFields_PRICE,
				shop.ArticleFields_PROMOTED,
			},
			&shop.ArticleRelations{
				Images: []shop.MediaFields{
					shop.MediaFields_MD_URL,
					shop.MediaFields_MD_LABEL,
				},
			},
			&shop.Limits{
				Limit: DefaultLimit,
			},
		},
		{
			"Custom",
			&shop.ListConditions{
				Fields: []shop.ArticleFields{shop.ArticleFields_TITLE},
				Relations: &shop.ArticleRelations{
					Videos: []shop.MediaFields{shop.MediaFields_MD_URL},
				},
				Limits: &shop.Limits{
					Limit:  99,
					Offset: 22,
				},
			},
			[]shop.ArticleFields{shop.ArticleFields_TITLE},
			&shop.ArticleRelations{
				Videos: []shop.MediaFields{shop.MediaFields_MD_URL},
			},
			&shop.Limits{
				Limit:  99,
				Offset: 22,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := defaults(tt.cond)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compat() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("compat() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("compat() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestArticleListQuery(t *testing.T) {
	type args struct {
		cond   *shop.ListConditions
		schema string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []interface{}
		wantErr bool
	}{
		{
			"Defaults",
			args{
				&shop.ListConditions{},
				"shop",
			},
			defaultArtQuery,
			nil,
			false,
		},
		{
			"All articles with images and categories",
			args{
				&shop.ListConditions{
					Fields: []shop.ArticleFields{
						shop.ArticleFields_ID,
						shop.ArticleFields_TITLE,
						shop.ArticleFields_PRICE,
					},
					Relations: &shop.ArticleRelations{
						Images: []shop.MediaFields{
							shop.MediaFields_MD_URL,
							shop.MediaFields_MD_LABEL,
						},
						Categories: []shop.CategoryFields{
							shop.CategoryFields_CAT_ID,
							shop.CategoryFields_CAT_LABEL,
						},
					},
				},
				"shop",
			},
			allArtsWithImagesAndCategories,
			nil,
			false,
		},
		{
			"Undefined field error",
			args{
				&shop.ListConditions{
					Fields: []shop.ArticleFields{
						99,
					},
				},
				"shop",
			},
			"",
			nil,
			true,
		},
		{
			"Undefined relation field error",
			args{
				&shop.ListConditions{
					Fields: []shop.ArticleFields{
						shop.ArticleFields_ID,
					},
					Relations: &shop.ArticleRelations{
						Images: []shop.MediaFields{
							99,
						},
					},
				},
				"shop",
			},
			"",
			nil,
			true,
		},
		{
			"All articles, all relations and all fields",
			args{
				&shop.ListConditions{
					Fields: []shop.ArticleFields{
						shop.ArticleFields_ALL,
					},
					Relations: &shop.ArticleRelations{
						Images:     []shop.MediaFields{shop.MediaFields_MD_ALL},
						Videos:     []shop.MediaFields{shop.MediaFields_MD_ALL},
						Categories: []shop.CategoryFields{shop.CategoryFields_CAT_ALL},
						Baseprices: []shop.BasePriceFields{shop.BasePriceFields_BP_ALL},
						Variants:   []shop.VariantFields{shop.VariantFields_VRT_ALL},
					},
				},
				"shop",
			},
			allArtsWithAllRelationsAndFields,
			nil,
			false,
		},
		{
			"Published articles in category with images",
			args{
				&shop.ListConditions{
					OnlyPublished:     true,
					OnlyCategoryLabel: "spanac",
					Fields: []shop.ArticleFields{
						shop.ArticleFields_ID,
						shop.ArticleFields_TITLE,
						shop.ArticleFields_PRICE,
					},
					Relations: &shop.ArticleRelations{
						Images: []shop.MediaFields{
							shop.MediaFields_MD_URL,
							shop.MediaFields_MD_LABEL,
						},
					},
				},
				"shop",
			},
			categoryPublishedArtsWithImages,
			[]interface{}{"spanac"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ArticleListQuery(tt.args.cond, tt.args.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleListQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ArticleListQuery() got = \n%v\nwant\n%v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ArticleListQuery() got1 = \n%v\nwant\n%v", got1, tt.want1)
			}
		})
	}
}

func BenchmarkArticleListQueryAllAll(b *testing.B) {
	Debug = false
	cond := &shop.ListConditions{
		Fields: []shop.ArticleFields{
			shop.ArticleFields_ALL,
		},
		Relations: &shop.ArticleRelations{
			Images:     []shop.MediaFields{shop.MediaFields_MD_ALL},
			Videos:     []shop.MediaFields{shop.MediaFields_MD_ALL},
			Categories: []shop.CategoryFields{shop.CategoryFields_CAT_ALL},
			Baseprices: []shop.BasePriceFields{shop.BasePriceFields_BP_ALL},
			Variants:   []shop.VariantFields{shop.VariantFields_VRT_ALL},
		},
	}

	for i := 0; i < b.N; i++ {
		ArticleListQuery(cond, "shop")
	}
}

func BenchmarkArticleListQueryImgCats(b *testing.B) {
	Debug = false
	cond := &shop.ListConditions{
		Fields: []shop.ArticleFields{
			shop.ArticleFields_ID,
			shop.ArticleFields_TITLE,
			shop.ArticleFields_PRICE,
		},
		Relations: &shop.ArticleRelations{
			Images: []shop.MediaFields{
				shop.MediaFields_MD_URL,
				shop.MediaFields_MD_LABEL,
			},
			Categories: []shop.CategoryFields{
				shop.CategoryFields_CAT_ID,
				shop.CategoryFields_CAT_LABEL,
			},
		},
	}

	for i := 0; i < b.N; i++ {
		ArticleListQuery(cond, "shop")
	}
}

func BenchmarkArticleListDefaults(b *testing.B) {
	Debug = false
	cond := &shop.ListConditions{}
	for i := 0; i < b.N; i++ {
		ArticleListQuery(cond, "shop")
	}
}

func BenchmarkArticleListWithFilters(b *testing.B) {
	Debug = false
	cond := &shop.ListConditions{
		OnlyPublished:     true,
		OnlyCategoryLabel: "spanac",
		Fields: []shop.ArticleFields{
			shop.ArticleFields_ID,
			shop.ArticleFields_TITLE,
			shop.ArticleFields_PRICE,
		},
		Relations: &shop.ArticleRelations{
			Images: []shop.MediaFields{
				shop.MediaFields_MD_URL,
				shop.MediaFields_MD_LABEL,
			},
		},
	}

	for i := 0; i < b.N; i++ {
		ArticleListQuery(cond, "shop")
	}
}

const (
	defaultArtQuery = `with filters as (
	select m.id
	from shop.articles m
	
),
arts as (
	select a.id, a.title, a.price, a.promoted
	from filters f
	join shop.articles a on a.id = f.id
	limit 25
	offset 0
),
r0 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'url', r.url, 'label', r.label
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.images r on r.article_id = arts.id
	group by arts.id
)
select json_agg(
	json_build_object(
		'id', a.id, 'title', a.title, 'price', a.price::text, 'promoted', a.promoted, 'images', r0.js
	)
)
from arts a
join r0 on a.id = r0.id;`

	allArtsWithAllRelationsAndFields = `with filters as (
	select m.id
	from shop.articles m
	
),
arts as (
	select a.id, a.created_at, a.updated_at, a.published, a.title, a.description, a.price, a.promoted
	from filters f
	join shop.articles a on a.id = f.id
	limit 25
	offset 0
),
r0 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'id', r.id, 'label', r.label, 'url', r.url
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.images r on r.article_id = arts.id
	group by arts.id
),
r1 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'id', r.id, 'label', r.label, 'url', r.url
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.videos r on r.article_id = arts.id
	group by arts.id
),
r2 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'id', r.id, 'created_at', r.created_at, 'updated_at', r.updated_at, 'label', r.label
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.category_articles j on j.article_id = arts.id
	left join shop.categories r on r.id = j.category_id
	group by arts.id
),
r3 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'id', r.id, 'created_at', r.created_at, 'updated_at', r.updated_at, 'label', r.label, 'price', r.price::text
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.article_base_prices j on j.article_id = arts.id
	left join shop.base_prices r on r.id = j.base_price_id
	group by arts.id
),
r4 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'id', r.id, 'created_at', r.created_at, 'updated_at', r.updated_at, 'labels', r.labels, 'multiplier', r.multiplier::text
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.variants r on r.article_id = arts.id
	group by arts.id
)
select json_agg(
	json_build_object(
		'id', a.id, 'created_at', a.created_at, 'updated_at', a.updated_at, 'published', a.published, 'title', a.title, 'description', a.description, 'price', a.price::text, 'promoted', a.promoted, 'images', r0.js, 'videos', r1.js, 'categories', r2.js, 'base_prices', r3.js, 'variants', r4.js
	)
)
from arts a
join r0 on a.id = r0.id
join r1 on a.id = r1.id
join r2 on a.id = r2.id
join r3 on a.id = r3.id
join r4 on a.id = r4.id;`

	categoryPublishedArtsWithImages = `with filters as (
	select m.id
	from shop.articles m
	join shop.category_articles ac on ac.article_id = m.id
	join shop.categories c on c.id = ac.category_id
	where m.published
	and c.label = $1
),
arts as (
	select a.id, a.title, a.price
	from filters f
	join shop.articles a on a.id = f.id
	limit 25
	offset 0
),
r0 as (
	select arts.id, coalesce(json_agg(
		json_build_object(
			'url', r.url, 'label', r.label
		)
	) filter (where r.id is not null), null::JSON) as js
	from arts
	left join shop.images r on r.article_id = arts.id
	group by arts.id
)
select json_agg(
	json_build_object(
		'id', a.id, 'title', a.title, 'price', a.price::text, 'images', r0.js
	)
)
from arts a
join r0 on a.id = r0.id;`

	allArtsOnlyOffset = `with filters as (
	select m.id, m.title, m.price
	from shop.articles m
	
),
arts as (
	select a.id, a.title, a.price
	from filters f
	join shop.articles a on a.id = f.id
	limit 25
	offset 100
)
select json_agg(
	json_build_object(
		'id', a.id, 'title', a.title, 'price', a.price::text
	)
)
from arts a
;`
)

func Test_filters(t *testing.T) {
	type args struct {
		cond   *shop.ListConditions
		schema string
	}
	tests := []struct {
		name        string
		args        args
		wantFilters string
		wantArgs    []interface{}
	}{
		{
			"Empty",
			args{
				&shop.ListConditions{},
				"shop",
			},
			"",
			nil,
		},
		{
			"Published",
			args{
				&shop.ListConditions{
					OnlyPublished: true,
				},
				"shop",
			},
			"\n\twhere m.published",
			nil,
		},
		{
			"Promoted",
			args{
				&shop.ListConditions{
					OnlyPromoted: true,
				},
				"shop",
			},
			"\n\twhere m.promoted",
			nil,
		},
		{
			"Category id",
			args{
				&shop.ListConditions{
					OnlyCategoryId: 5,
				},
				"shop",
			},
			`join shop.category_articles ac on ac.article_id = m.id
	where ac.category_id = $1`,
			[]interface{}{int32(5)},
		},
		{
			"Category id",
			args{
				&shop.ListConditions{
					OnlyCategoryLabel: "spanac",
				},
				"shop",
			},
			`join shop.category_articles ac on ac.article_id = m.id
	join shop.categories c on c.id = ac.category_id
	where c.label = $1`,
			[]interface{}{"spanac"},
		},
		{
			"Published and category id",
			args{
				&shop.ListConditions{
					OnlyPublished:     true,
					OnlyCategoryLabel: "spanac",
				},
				"shop",
			},
			`join shop.category_articles ac on ac.article_id = m.id
	join shop.categories c on c.id = ac.category_id
	where m.published
	and c.label = $1`,
			[]interface{}{"spanac"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilters, gotArgs := filters(tt.args.cond, tt.args.schema)
			if gotFilters != tt.wantFilters {
				t.Errorf("filters() gotFilters = %v, want %v", gotFilters, tt.wantFilters)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("filters() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}
