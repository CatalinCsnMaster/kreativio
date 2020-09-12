// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

package builder

import (
	"sort"

	"github.com/moapis/shop"
	"github.com/moapis/shop/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	fieldErr = "Field %s not defined"
)

type fieldColumns struct {
	M   map[int]string
	all []string
}

func (fc fieldColumns) columns(fids []int) ([]string, int) {
	cols := make([]string, len(fids))
	for i, id := range fids {
		if id == 0 { // 0 is used for *_ALL
			return fc.all, 0
		}

		col, ok := fc.M[id]
		if !ok {
			return nil, id
		}
		cols[i] = col
	}
	return cols, 0
}

var (
	// ArticleFieldColumns maps requested article fields to columns
	ArticleFieldColumns = fieldColumns{
		M: map[int]string{
			int(shop.ArticleFields_ID):          models.ArticleColumns.ID,
			int(shop.ArticleFields_CREATED):     models.ArticleColumns.CreatedAt,
			int(shop.ArticleFields_UPDATED):     models.ArticleColumns.UpdatedAt,
			int(shop.ArticleFields_PUBLISHED):   models.ArticleColumns.Published,
			int(shop.ArticleFields_TITLE):       models.ArticleColumns.Title,
			int(shop.ArticleFields_DESCRIPTION): models.ArticleColumns.Description,
			int(shop.ArticleFields_PRICE):       models.ArticleColumns.Price,
			int(shop.ArticleFields_PROMOTED):    models.ArticleColumns.Promoted,
		},
	}

	// ImageFieldColumns maps requested image fields to columns
	ImageFieldColumns = fieldColumns{
		M: map[int]string{
			int(shop.MediaFields_MD_ID):    models.ImageColumns.ID,
			int(shop.MediaFields_MD_LABEL): models.ImageColumns.Label,
			int(shop.MediaFields_MD_URL):   models.ImageColumns.URL,
		},
	}

	// VideoFieldColumns maps requested video fields to columns
	VideoFieldColumns = fieldColumns{
		M: map[int]string{
			int(shop.MediaFields_MD_ID):    models.VideoColumns.ID,
			int(shop.MediaFields_MD_LABEL): models.VideoColumns.Label,
			int(shop.MediaFields_MD_URL):   models.VideoColumns.URL,
		},
	}

	// CategoryFieldColumns maps requested category fields to columns
	CategoryFieldColumns = fieldColumns{
		M: map[int]string{
			int(shop.CategoryFields_CAT_ID):      models.CategoryColumns.ID,
			int(shop.CategoryFields_CAT_CREATED): models.CategoryColumns.CreatedAt,
			int(shop.CategoryFields_CAT_UPDATED): models.CategoryColumns.UpdatedAt,
			int(shop.CategoryFields_CAT_LABEL):   models.CategoryColumns.Label,
		}}

	// BasePriceFieldColumns maps requested baseprice fields to columns
	BasePriceFieldColumns = fieldColumns{
		M: map[int]string{
			int(shop.BasePriceFields_BP_ID):      models.BasePriceColumns.ID,
			int(shop.BasePriceFields_BP_CREATED): models.BasePriceColumns.CreatedAt,
			int(shop.BasePriceFields_BP_UPDATED): models.BasePriceColumns.UpdatedAt,
			int(shop.BasePriceFields_BP_LABEL):   models.BasePriceColumns.Label,
			int(shop.BasePriceFields_BP_PRICE):   models.BasePriceColumns.Price,
		},
	}

	// VariantFieldColumns maps requested variant fields to columns
	VariantFieldColumns = fieldColumns{
		M: map[int]string{
			int(shop.VariantFields_VRT_ID):         models.VariantColumns.ID,
			int(shop.VariantFields_VRT_CREATED):    models.VariantColumns.CreatedAt,
			int(shop.VariantFields_VRT_UPDATED):    models.VariantColumns.UpdatedAt,
			int(shop.VariantFields_VRT_LABELS):     models.VariantColumns.Labels,
			int(shop.VariantFields_VRT_MULTIPLIER): models.VariantColumns.Multiplier,
		},
	}
)

// init the "all" fields
func init() {
	fcs := []*fieldColumns{
		&ArticleFieldColumns,
		&ImageFieldColumns,
		&VideoFieldColumns,
		&CategoryFieldColumns,
		&BasePriceFieldColumns,
		&VariantFieldColumns,
	}

	for _, fc := range fcs {
		// sort for deterministic output
		s := make([]int, 0, len(fc.M))
		for k := range fc.M {
			s = append(s, k)
		}
		sort.Ints(s)

		cols := make([]string, len(s))
		for i := 0; i < len(s); i++ {
			cols[i] = fc.M[s[i]]
		}
		fc.all = cols
	}
}

// ArticleColumns returns a list of columns, reflecting the requested fields.
// When shop.ArticleFields_ALL is passed, all columns are returned
// In case of an unmapped field, codes.Unimplemented error is returned.
func ArticleColumns(fields []shop.ArticleFields) ([]string, error) {
	fids := make([]int, len(fields))
	for i, f := range fields {
		fids[i] = int(f)
	}

	cols, id := ArticleFieldColumns.columns(fids)
	if id != 0 {
		return nil, status.Errorf(codes.Unimplemented, fieldErr, shop.ArticleFields(id))
	}

	return cols, nil
}

// ImageColumns returns a list of columns, reflecting the requested fields.
// When shop.MediaFields_MD_ALL is passed, all columns are returned
// In case of an unmapped field, codes.Unimplemented error is returned.
func ImageColumns(fields []shop.MediaFields) ([]string, error) {
	fids := make([]int, len(fields))
	for i, f := range fields {
		fids[i] = int(f)
	}

	cols, id := ImageFieldColumns.columns(fids)
	if id != 0 {
		return nil, status.Errorf(codes.Unimplemented, fieldErr, shop.MediaFields(id))
	}

	return cols, nil
}

// VideoColumns returns a list of columns, reflecting the requested fields.
// When shop.MediaFields_MD_ALL is passed, all columns are returned
// In case of an unmapped field, codes.Unimplemented error is returned.
func VideoColumns(fields []shop.MediaFields) ([]string, error) {
	fids := make([]int, len(fields))
	for i, f := range fields {
		fids[i] = int(f)
	}

	cols, id := VideoFieldColumns.columns(fids)
	if id != 0 {
		return nil, status.Errorf(codes.Unimplemented, fieldErr, shop.MediaFields(id))
	}

	return cols, nil
}

// CategoryColumns returns a list of columns, reflecting the requested fields.
// When shop.CategoryFields_CAT_ALL is passed, all columns are returned
// In case of an unmapped field, codes.Unimplemented error is returned.
func CategoryColumns(fields []shop.CategoryFields) ([]string, error) {
	fids := make([]int, len(fields))
	for i, f := range fields {
		fids[i] = int(f)
	}

	cols, id := CategoryFieldColumns.columns(fids)
	if id != 0 {
		return nil, status.Errorf(codes.Unimplemented, fieldErr, shop.CategoryFields(id))
	}

	return cols, nil
}

// BasePriceColumns returns a list of columns, reflecting the requested fields.
// When shop.BasePriceFields_BP_ALL is passed, all columns are returned
// In case of an unmapped field, codes.Unimplemented error is returned.
func BasePriceColumns(fields []shop.BasePriceFields) ([]string, error) {
	fids := make([]int, len(fields))
	for i, f := range fields {
		fids[i] = int(f)
	}

	cols, id := BasePriceFieldColumns.columns(fids)
	if id != 0 {
		return nil, status.Errorf(codes.Unimplemented, fieldErr, shop.BasePriceFields(id))
	}

	return cols, nil
}

// VariantColumns returns a list of columns, reflecting the requested fields.
// When shop.VariantFields_VRT_ALL is passed, all columns are returned
// In case of an unmapped field, codes.Unimplemented error is returned.
func VariantColumns(fields []shop.VariantFields) ([]string, error) {
	fids := make([]int, len(fields))
	for i, f := range fields {
		fids[i] = int(f)
	}

	cols, id := VariantFieldColumns.columns(fids)
	if id != 0 {
		return nil, status.Errorf(codes.Unimplemented, fieldErr, shop.VariantFields(id))
	}

	return cols, nil
}

const schema = "shop"

func articleRelations(rel *shop.ArticleRelations) ([]relation, error) {
	var (
		br   = make([]relation, 0, 5)
		cols []string
		errs [5]error
	)

	cols, errs[0] = ImageColumns(rel.GetImages())
	if len(cols) > 0 {
		br = append(br, relation{
			name:    models.TableNames.Images,
			columns: cols,
			id:      models.ImageColumns.ArticleID,
		})
	}

	cols, errs[1] = VideoColumns(rel.GetVideos())
	if len(cols) > 0 {
		br = append(br, relation{
			name:    models.TableNames.Videos,
			columns: cols,
			id:      models.VideoColumns.ArticleID,
		})
	}

	cols, errs[2] = CategoryColumns(rel.GetCategories())
	if len(cols) > 0 {
		br = append(br, relation{
			name:      models.TableNames.Categories,
			columns:   cols,
			id:        models.CategoryColumns.ID,
			joinTable: models.TableNames.CategoryArticles,
			joinIDs:   [2]string{"article_id", "category_id"},
		})
	}

	cols, errs[3] = BasePriceColumns(rel.GetBaseprices())
	if len(cols) > 0 {
		br = append(br, relation{
			name:      models.TableNames.BasePrices,
			columns:   cols,
			id:        models.BasePriceColumns.ID,
			joinTable: models.TableNames.ArticleBasePrices,
			joinIDs:   [2]string{"article_id", "base_price_id"},
		})
	}

	cols, errs[4] = VariantColumns(rel.GetVariants())
	if len(cols) > 0 {
		br = append(br, relation{
			name:    models.TableNames.Variants,
			columns: cols,
			id:      models.VariantColumns.ArticleID,
		})
	}

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return br, nil
}
