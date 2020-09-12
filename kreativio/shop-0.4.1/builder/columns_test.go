// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

package builder

import (
	"errors"
	"reflect"
	"testing"

	"github.com/moapis/shop"
	"github.com/moapis/shop/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_fieldColumns_columns(t *testing.T) {
	tfc := fieldColumns{
		M: map[int]string{
			1: "one",
			2: "two",
			3: "three",
		},
		all: []string{"one", "two", "three"},
	}

	tests := []struct {
		name  string
		fids  []int
		want  []string
		want1 int
	}{
		{
			"Nil gives emtpy",
			nil,
			[]string{},
			0,
		},
		{
			"len == 0 gives empty",
			[]int{},
			[]string{},
			0,
		},
		{
			"2 and 3",
			[]int{2, 3},
			[]string{"two", "three"},
			0,
		},
		{
			"Entry with 0 returns all",
			[]int{2, 0},
			[]string{"one", "two", "three"},
			0,
		},
		{
			"Non existing entry returns fid number",
			[]int{2, 3, 4},
			nil,
			4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tfc.columns(tt.fids)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fieldColumns.columns() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("fieldColumns.columns() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestArticleColumns(t *testing.T) {
	tests := []struct {
		name    string
		fields  []shop.ArticleFields
		want    []string
		wantErr error
	}{
		{
			"Nil returns emtpy",
			nil,
			[]string{},
			nil,
		},
		{
			"Some columns",
			[]shop.ArticleFields{
				shop.ArticleFields_ID,
				shop.ArticleFields_TITLE,
				shop.ArticleFields_PRICE,
			},
			[]string{
				models.ArticleColumns.ID,
				models.ArticleColumns.Title,
				models.ArticleColumns.Price,
			},
			nil,
		},
		{
			"With all returns all",
			[]shop.ArticleFields{
				shop.ArticleFields_ID,
				shop.ArticleFields_TITLE,
				shop.ArticleFields_ALL,
			},
			ArticleFieldColumns.all,
			nil,
		},
		{
			"Unknown ID error",
			[]shop.ArticleFields{
				shop.ArticleFields_ID,
				shop.ArticleFields_TITLE,
				shop.ArticleFields(9),
			},
			nil,
			status.Errorf(codes.Unimplemented, fieldErr, shop.ArticleFields(9)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ArticleColumns(tt.fields)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ArticleColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArticleColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageColumns(t *testing.T) {
	tests := []struct {
		name    string
		fields  []shop.MediaFields
		want    []string
		wantErr error
	}{
		{
			"Nil returns emtpy",
			nil,
			[]string{},
			nil,
		},
		{
			"Some columns",
			[]shop.MediaFields{
				shop.MediaFields_MD_LABEL,
				shop.MediaFields_MD_URL,
			},
			[]string{
				models.ImageColumns.Label,
				models.ImageColumns.URL,
			},
			nil,
		},
		{
			"With all returns all",
			[]shop.MediaFields{
				shop.MediaFields_MD_LABEL,
				shop.MediaFields_MD_ALL,
			},
			ImageFieldColumns.all,
			nil,
		},
		{
			"Unknown ID error",
			[]shop.MediaFields{
				shop.MediaFields_MD_LABEL,
				shop.MediaFields_MD_URL,
				shop.MediaFields(9),
			},
			nil,
			status.Errorf(codes.Unimplemented, fieldErr, shop.MediaFields(9)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ImageColumns(tt.fields)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ArticleColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArticleColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVideoColumns(t *testing.T) {
	tests := []struct {
		name    string
		fields  []shop.MediaFields
		want    []string
		wantErr error
	}{
		{
			"Nil returns emtpy",
			nil,
			[]string{},
			nil,
		},
		{
			"Some columns",
			[]shop.MediaFields{
				shop.MediaFields_MD_LABEL,
				shop.MediaFields_MD_URL,
			},
			[]string{
				models.VideoColumns.Label,
				models.VideoColumns.URL,
			},
			nil,
		},
		{
			"With all returns all",
			[]shop.MediaFields{
				shop.MediaFields_MD_LABEL,
				shop.MediaFields_MD_ALL,
			},
			VideoFieldColumns.all,
			nil,
		},
		{
			"Unknown ID error",
			[]shop.MediaFields{
				shop.MediaFields_MD_LABEL,
				shop.MediaFields_MD_URL,
				shop.MediaFields(9),
			},
			nil,
			status.Errorf(codes.Unimplemented, fieldErr, shop.MediaFields(9)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VideoColumns(tt.fields)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ArticleColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArticleColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategoryColumns(t *testing.T) {
	tests := []struct {
		name    string
		fields  []shop.CategoryFields
		want    []string
		wantErr error
	}{
		{
			"Nil returns emtpy",
			nil,
			[]string{},
			nil,
		},
		{
			"Some columns",
			[]shop.CategoryFields{
				shop.CategoryFields_CAT_ID,
				shop.CategoryFields_CAT_LABEL,
			},
			[]string{
				models.CategoryColumns.ID,
				models.CategoryColumns.Label,
			},
			nil,
		},
		{
			"With all returns all",
			[]shop.CategoryFields{
				shop.CategoryFields_CAT_ID,
				shop.CategoryFields_CAT_ALL,
			},
			CategoryFieldColumns.all,
			nil,
		},
		{
			"Unknown ID error",
			[]shop.CategoryFields{
				shop.CategoryFields_CAT_LABEL,
				shop.CategoryFields(9),
			},
			nil,
			status.Errorf(codes.Unimplemented, fieldErr, shop.CategoryFields(9)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CategoryColumns(tt.fields)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ArticleColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArticleColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBasePriceColumns(t *testing.T) {
	tests := []struct {
		name    string
		fields  []shop.BasePriceFields
		want    []string
		wantErr error
	}{
		{
			"Nil returns emtpy",
			nil,
			[]string{},
			nil,
		},
		{
			"Some columns",
			[]shop.BasePriceFields{
				shop.BasePriceFields_BP_ID,
				shop.BasePriceFields_BP_LABEL,
			},
			[]string{
				models.BasePriceColumns.ID,
				models.BasePriceColumns.Label,
			},
			nil,
		},
		{
			"With all returns all",
			[]shop.BasePriceFields{
				shop.BasePriceFields_BP_ID,
				shop.BasePriceFields_BP_ALL,
			},
			BasePriceFieldColumns.all,
			nil,
		},
		{
			"Unknown ID error",
			[]shop.BasePriceFields{
				shop.BasePriceFields_BP_ID,
				shop.BasePriceFields(9),
			},
			nil,
			status.Errorf(codes.Unimplemented, fieldErr, shop.MediaFields(9)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BasePriceColumns(tt.fields)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ArticleColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArticleColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariantColumns(t *testing.T) {
	tests := []struct {
		name    string
		fields  []shop.VariantFields
		want    []string
		wantErr error
	}{
		{
			"Nil returns emtpy",
			nil,
			[]string{},
			nil,
		},
		{
			"Some columns",
			[]shop.VariantFields{
				shop.VariantFields_VRT_ID,
				shop.VariantFields_VRT_LABELS,
			},
			[]string{
				models.VariantColumns.ID,
				models.VariantColumns.Labels,
			},
			nil,
		},
		{
			"With all returns all",
			[]shop.VariantFields{
				shop.VariantFields_VRT_ID,
				shop.VariantFields_VRT_ALL,
			},
			VariantFieldColumns.all,
			nil,
		},
		{
			"Unknown ID error",
			[]shop.VariantFields{
				shop.VariantFields_VRT_ID,
				shop.VariantFields(9),
			},
			nil,
			status.Errorf(codes.Unimplemented, fieldErr, shop.MediaFields(9)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VariantColumns(tt.fields)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("VariantColumns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VariantColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_articleRelations(t *testing.T) {
	tests := []struct {
		name    string
		rel     *shop.ArticleRelations
		want    []relation
		wantErr bool
	}{
		{
			"Image label",
			&shop.ArticleRelations{
				Images: []shop.MediaFields{shop.MediaFields_MD_LABEL},
			},
			[]relation{{
				name:    models.TableNames.Images,
				columns: []string{models.ImageColumns.Label},
				id:      models.ImageColumns.ArticleID,
			}},
			false,
		},
		{
			"Video label",
			&shop.ArticleRelations{
				Videos: []shop.MediaFields{shop.MediaFields_MD_LABEL},
			},
			[]relation{{
				name:    models.TableNames.Videos,
				columns: []string{models.ImageColumns.Label},
				id:      models.VideoColumns.ArticleID,
			}},
			false,
		},
		{
			"Category label",
			&shop.ArticleRelations{
				Categories: []shop.CategoryFields{shop.CategoryFields_CAT_LABEL},
			},
			[]relation{{
				name:      "categories",
				columns:   []string{models.CategoryColumns.Label},
				id:        models.CategoryColumns.ID,
				joinTable: models.TableNames.CategoryArticles,
				joinIDs:   [2]string{"article_id", "category_id"},
			}},
			false,
		},
		{
			"Baseprice label",
			&shop.ArticleRelations{
				Baseprices: []shop.BasePriceFields{shop.BasePriceFields_BP_LABEL},
			},
			[]relation{{
				name:      "base_prices",
				columns:   []string{models.BasePriceColumns.Label},
				id:        models.BasePriceColumns.ID,
				joinTable: models.TableNames.ArticleBasePrices,
				joinIDs:   [2]string{"article_id", "base_price_id"},
			}},
			false,
		},
		{
			"Variant labels",
			&shop.ArticleRelations{
				Variants: []shop.VariantFields{shop.VariantFields_VRT_LABELS},
			},
			[]relation{{
				name:    "variants",
				columns: []string{models.VariantColumns.Labels},
				id:      models.VariantColumns.ArticleID,
			}},
			false,
		},
		{
			"An error",
			&shop.ArticleRelations{
				Images: []shop.MediaFields{99},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := articleRelations(tt.rel)
			if (err != nil) != tt.wantErr {
				t.Errorf("articleRelations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("articleRelations() = %v, want %v", got, tt.want)
			}
		})
	}
}
