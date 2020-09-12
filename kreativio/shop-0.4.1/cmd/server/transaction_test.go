// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/moapis/shop"
	"github.com/moapis/shop/mobilpay"
	"github.com/moapis/shop/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func Test_shopServer_newTx(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx      context.Context
		method   string
		readOnly bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Success",
			args{
				context.Background(),
				"testing",
				false,
			},
			false,
		},
		{
			"Context error",
			args{
				ectx,
				"testing",
				false,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.newTx(tt.args.ctx, tt.args.method, tt.args.readOnly)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.newTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("shopServer.newTx() = %v, want %v", got, "something")
			}
		})
	}
}

func Test_requestTx_upsertArticle(t *testing.T) {
	tests := []struct {
		name    string
		as      *shop.Article
		want    *models.Article
		wantErr bool
	}{
		{
			"New article",
			&shop.Article{
				Published:   true,
				Title:       "New article",
				Description: "Some stuff",
				Price:       "49.98",
				Promoted:    true,
			},
			&models.Article{
				ID:          1,
				Published:   true,
				Title:       "New article",
				Description: "Some stuff",
				Price:       types.NewDecimal(decimal.New(4998, 2)), // 49.98
				Promoted:    true,
			},
			false,
		},
		{
			"Update article",
			&shop.Article{
				Id:          11, //This test article starts unpublished
				Published:   true,
				Title:       "ID 11+",
				Description: "This is the first article, published",
				Price:       "31000.10",
				Promoted:    true,
			},
			&models.Article{
				ID:          11,
				Published:   true,
				Title:       "ID 11+",
				Description: "This is the first article, published",
				Price:       types.NewDecimal(decimal.New(3100010, 2)), // 31000.10
				Promoted:    true,
			},
			false,
		},
		{
			"Insert err",
			&shop.Article{
				Published:   true,
				Title:       "Error article",
				Description: "Some stuff",
				Price:       "49.98",
			},
			nil,
			true,
		},
		{
			"Missing fields",
			&shop.Article{},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.wantErr {
				rt.Done()
			}

			got, err := rt.upsertArticle(tt.as)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.upsertArticle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (tt.want.ID != got.ID || tt.want.Published != got.Published || tt.want.Title != got.Title || tt.want.Description != got.Description || tt.want.Price.String() != got.Price.String()) {
				t.Errorf("requestTx.upsertArticle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_setArticleCategories(t *testing.T) {
	type args struct {
		art *models.Article
		sc  []*shop.Category
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.Category
		wantErr bool
	}{
		{
			"Success",
			args{
				testArticles[1],
				[]*shop.Category{
					{Id: 20},
					{Id: 22},
				},
			},
			[]*models.Category{
				{ID: 20},
				{ID: 22},
			},
			false,
		},
		{
			"Zero category ID",
			args{
				testArticles[1],
				[]*shop.Category{
					{Id: 0},
					{Id: 22},
				},
			},
			nil,
			true,
		},
		{
			"Non-existing category ID",
			args{
				testArticles[1],
				[]*shop.Category{
					{Id: 999},
					{Id: 22},
				},
			},
			nil,
			true,
		},
		{
			"No categories",
			args{
				testArticles[1],
				nil,
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()

			if err := rt.setArticleCategories(tt.args.art, tt.args.sc); (err != nil) != tt.wantErr {
				t.Fatalf("requestTx.setArticleCategories() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := tt.args.art.Categories().All(rt.Ctx, rt.Tx)
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != len(tt.want) {
					t.Fatalf("requestTx.setArticleCategories() = %v, want %v", got, tt.want)
				}
				for i, w := range tt.want {
					if got[i].ID != w.ID {
						t.Errorf("requestTx.setArticleCategories() = \n%v\nwant\n%v", got[i], w)
					}
				}
			}
		})
	}
}

var (
	testShopBasePrices = []*shop.BasePrice{
		{
			Id:      31,
			Created: &timestamp.Timestamp{Seconds: 4000},
			Updated: &timestamp.Timestamp{Seconds: 5000},
			Label:   "Cheap material",
			Price:   "44.55",
		},
		{
			Id:      32,
			Created: &timestamp.Timestamp{Seconds: 6000},
			Updated: &timestamp.Timestamp{Seconds: 7000},
			Label:   "Premium material",
			Price:   "99.99",
		},
	}
)

func Test_requestTx_setArticleBasePrices(t *testing.T) {
	art := *testArticles[0]

	mid := proto.Clone(testShopBasePrices[0]).(*shop.BasePrice)
	mid.Id = 0

	type args struct {
		art *models.Article
		sbp []*shop.BasePrice
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.BasePrice
		wantErr bool
	}{
		{
			"Success",
			args{
				&art,
				testShopBasePrices,
			},
			testBasePrices,
			false,
		},
		{
			"Missing Id",
			args{
				&art,
				[]*shop.BasePrice{mid},
			},
			nil,
			true,
		},
		{
			"DB Error",
			args{
				&art,
				testShopBasePrices,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			if err := rt.setArticleBasePrices(tt.args.art, tt.args.sbp); (err != nil) != tt.wantErr {
				t.Errorf("requestTx.setArticleBasePrices() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := art.BasePrices().All(rt.Ctx, rt.Tx)
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != len(tt.want) {
					t.Fatalf("requestTx.setArticleBasePrices() = %v, want %v", got, tt.want)
				}
				for i, w := range tt.want {
					if got[i].ID != w.ID {
						t.Errorf("requestTx.setArticleBasePrices() = \n%v\nwant\n%v", got[i], w)
					}
				}
			}
		})
	}
}

var (
	testShopVariants = []*shop.Variant{
		{
			Id:         41,
			Labels:     []string{"hello", "world"},
			Multiplier: "3.33",
		},
		{
			Id:         42,
			Labels:     []string{"foo", "bar"},
			Multiplier: "9.99",
		},
	}
)

func Test_requestTx_updateVariants(t *testing.T) {
	type args struct {
		aid int
		sv  []*shop.Variant
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.Variant
		wantErr error
	}{
		{
			"Same variants",
			args{
				13,
				testShopVariants,
			},
			testVariants,
			nil,
		},
		{
			"All new variant",
			args{
				13,
				[]*shop.Variant{
					{
						Labels:     []string{"a", "new", "hope"},
						Multiplier: "2",
					},
				},
			},
			[]*models.Variant{
				{
					ID:         1,
					ArticleID:  13,
					Labels:     []string{"a", "new", "hope"},
					Multiplier: types.NewDecimal(decimal.New(2, 0)),
				},
			},
			nil,
		},
		{
			"Invalid multiplier",
			args{
				13,
				[]*shop.Variant{
					{
						Labels:     []string{"a", "new", "hope"},
						Multiplier: "spanac",
					},
				},
			},
			nil,
			status.Errorf(codes.InvalidArgument, errDecimal, varDecimal, "spanac"),
		},
		{
			"Cleanup Error",
			args{
				13,
				[]*shop.Variant{
					{
						Labels:     []string{"a", "new", "hope"},
						Multiplier: "2",
					},
				},
			},
			nil,
			status.Error(codes.Internal, errDB),
		},
		{
			"Non-existing article",
			args{
				31,
				[]*shop.Variant{
					{
						Labels:     []string{"a", "new", "hope"},
						Multiplier: "2",
					},
				},
			},
			nil,
			status.Error(codes.Internal, errDB),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "Cleanup Error" {
				rt.Done()
			}

			if err := rt.updateVariants(tt.args.aid, tt.args.sv); !errors.Is(err, tt.wantErr) {
				t.Errorf("requestTx.updateVariants() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				got, err := models.Variants(models.VariantWhere.ArticleID.EQ(tt.args.aid)).All(rt.Ctx, rt.Tx)
				if len(tt.want) != len(got) {
					t.Errorf("requestTx.updateVariants() = %v, want %v", got, tt.want)
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				for k, w := range tt.want {
					if w.ID != got[k].ID || w.ArticleID != got[k].ArticleID || !reflect.DeepEqual(w.Labels, got[k].Labels) {
						t.Errorf("requestTx.updateVariants() = \n%v\nwant\n%v", got[k], w)
					}
				}
			}
		})
	}
}

func Test_requestTx_updateImages(t *testing.T) {
	type args struct {
		aid int
		mds []*shop.Media
	}
	tests := []struct {
		name    string
		args    args
		want    models.ImageSlice
		wantErr bool
	}{
		{
			"Re-order",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First image on second article",
						Url:   "https://bucket.s3.com/firstonsecond.jpg",
					},
					{
						Id:    123,
						Label: "Third image on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.jpg",
					},
					{
						Id:    122,
						Label: "Second image on second article",
						Url:   "https://bucket.s3.com/secondonsecond.jpg",
					},
				},
			},
			[]*models.Image{
				{
					ID:        121,
					ArticleID: 12,
					Position:  1,
					Label:     "First image on second article",
					URL:       "https://bucket.s3.com/firstonsecond.jpg",
				},
				{
					ID:        123,
					ArticleID: 12,
					Position:  2,
					Label:     "Third image on second article",
					URL:       "https://bucket.s3.com/thirdonsecond.jpg",
				},
				{
					ID:        122,
					ArticleID: 12,
					Position:  3,
					Label:     "Second image on second article",
					URL:       "https://bucket.s3.com/secondonsecond.jpg",
				},
			},
			false,
		},
		{
			"With new image",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First image on second article",
						Url:   "https://bucket.s3.com/firstonsecond.jpg",
					},
					{
						Id:    123,
						Label: "Third image on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.jpg",
					},
					{
						Id:    122,
						Label: "Second image on second article",
						Url:   "https://bucket.s3.com/secondonsecond.jpg",
					},
					{
						Label: "New image",
						Url:   "https://bucket.s3.com/new.jpg",
					},
				},
			},
			[]*models.Image{
				{
					ID:        121,
					ArticleID: 12,
					Position:  1,
					Label:     "First image on second article",
					URL:       "https://bucket.s3.com/firstonsecond.jpg",
				},
				{
					ID:        123,
					ArticleID: 12,
					Position:  2,
					Label:     "Third image on second article",
					URL:       "https://bucket.s3.com/thirdonsecond.jpg",
				},
				{
					ID:        122,
					ArticleID: 12,
					Position:  3,
					Label:     "Second image on second article",
					URL:       "https://bucket.s3.com/secondonsecond.jpg",
				},
				{
					ID:        1,
					ArticleID: 12,
					Position:  4,
					Label:     "New image",
					URL:       "https://bucket.s3.com/new.jpg",
				},
			},
			false,
		},
		{
			"Removed image",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First image on second article",
						Url:   "https://bucket.s3.com/firstonsecond.jpg",
					},
					{
						Id:    123,
						Label: "Third image on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.jpg",
					},
				},
			},
			[]*models.Image{
				{
					ID:        121,
					ArticleID: 12,
					Position:  1,
					Label:     "First image on second article",
					URL:       "https://bucket.s3.com/firstonsecond.jpg",
				},
				{
					ID:        123,
					ArticleID: 12,
					Position:  2,
					Label:     "Third image on second article",
					URL:       "https://bucket.s3.com/thirdonsecond.jpg",
				},
			},
			false,
		},
		{
			"TX error",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First image on second article",
						Url:   "https://bucket.s3.com/firstonsecond.jpg",
					},
					{
						Id:    123,
						Label: "Third image on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.jpg",
					},
				},
			},
			nil,
			true,
		},
		{
			"Non existing article",
			args{
				99,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First image on second article",
						Url:   "https://bucket.s3.com/firstonsecond.jpg",
					},
					{
						Id:    123,
						Label: "Third image on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.jpg",
					},
				},
			},
			nil,
			true,
		},
		{
			"Emtpy fields",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "",
						Url:   "https://bucket.s3.com/firstonsecond.jpg",
					},
					{
						Id:    123,
						Label: "Third image on second article",
						Url:   "",
					},
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "TX error" {
				rt.Done()
			}

			if err := rt.updateImages(tt.args.aid, tt.args.mds); (err != nil) != tt.wantErr {
				t.Errorf("requestTx.updateImages() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := models.Images(models.ImageWhere.ArticleID.EQ(tt.args.aid), qm.OrderBy(models.ImageColumns.Position)).All(rt.Ctx, rt.Tx)
				if len(tt.want) != len(got) {
					t.Errorf("requestTx.updateImages() = %v, want %v", got, tt.want)
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				for k, w := range tt.want {
					if w.ID != got[k].ID || w.ArticleID != got[k].ArticleID || w.Position != got[k].Position || w.Label != got[k].Label || w.URL != got[k].URL {
						t.Errorf("requestTx.updateImages() = \n%v\nwant\n%v", got[k], w)
					}
				}
				// check if we did not delete other images
				img, err := models.FindImage(rt.Ctx, rt.Tx, testImages[0].ID, models.ImageColumns.ID)
				if err != nil || img.ID != testImages[0].ID {
					t.Errorf("requestTx.updateImages() = accidental delete")
				}
			}
		})
	}
}

func Test_requestTx_updateVideos(t *testing.T) {
	type args struct {
		aid int
		mds []*shop.Media
	}
	tests := []struct {
		name    string
		args    args
		want    models.VideoSlice
		wantErr bool
	}{
		{
			"Re-order",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First video on second article",
						Url:   "https://bucket.s3.com/firstonsecond.mp4",
					},
					{
						Id:    123,
						Label: "Third video on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.mp4",
					},
					{
						Id:    122,
						Label: "Second video on second article",
						Url:   "https://bucket.s3.com/secondonsecond.mp4",
					},
				},
			},
			[]*models.Video{
				{
					ID:        121,
					ArticleID: 12,
					Position:  1,
					Label:     "First video on second article",
					URL:       "https://bucket.s3.com/firstonsecond.mp4",
				},
				{
					ID:        123,
					ArticleID: 12,
					Position:  2,
					Label:     "Third video on second article",
					URL:       "https://bucket.s3.com/thirdonsecond.mp4",
				},
				{
					ID:        122,
					ArticleID: 12,
					Position:  3,
					Label:     "Second video on second article",
					URL:       "https://bucket.s3.com/secondonsecond.mp4",
				},
			},
			false,
		},
		{
			"With new video",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First video on second article",
						Url:   "https://bucket.s3.com/firstonsecond.mp4",
					},
					{
						Id:    123,
						Label: "Third video on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.mp4",
					},
					{
						Id:    122,
						Label: "Second video on second article",
						Url:   "https://bucket.s3.com/secondonsecond.mp4",
					},
					{
						Label: "New video",
						Url:   "https://bucket.s3.com/new.mp4",
					},
				},
			},
			[]*models.Video{
				{
					ID:        121,
					ArticleID: 12,
					Position:  1,
					Label:     "First video on second article",
					URL:       "https://bucket.s3.com/firstonsecond.mp4",
				},
				{
					ID:        123,
					ArticleID: 12,
					Position:  2,
					Label:     "Third video on second article",
					URL:       "https://bucket.s3.com/thirdonsecond.mp4",
				},
				{
					ID:        122,
					ArticleID: 12,
					Position:  3,
					Label:     "Second video on second article",
					URL:       "https://bucket.s3.com/secondonsecond.mp4",
				},
				{
					ID:        1,
					ArticleID: 12,
					Position:  4,
					Label:     "New video",
					URL:       "https://bucket.s3.com/new.mp4",
				},
			},
			false,
		},
		{
			"Removed video",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First video on second article",
						Url:   "https://bucket.s3.com/firstonsecond.mp4",
					},
					{
						Id:    123,
						Label: "Third video on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.mp4",
					},
				},
			},
			[]*models.Video{
				{
					ID:        121,
					ArticleID: 12,
					Position:  1,
					Label:     "First video on second article",
					URL:       "https://bucket.s3.com/firstonsecond.mp4",
				},
				{
					ID:        123,
					ArticleID: 12,
					Position:  2,
					Label:     "Third video on second article",
					URL:       "https://bucket.s3.com/thirdonsecond.mp4",
				},
			},
			false,
		},
		{
			"TX error",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First video on second article",
						Url:   "https://bucket.s3.com/firstonsecond.mp4",
					},
					{
						Id:    123,
						Label: "Third video on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.mp4",
					},
				},
			},
			nil,
			true,
		},
		{
			"Non existing article",
			args{
				99,
				[]*shop.Media{
					{
						Id:    121,
						Label: "First video on second article",
						Url:   "https://bucket.s3.com/firstonsecond.mp4",
					},
					{
						Id:    123,
						Label: "Third video on second article",
						Url:   "https://bucket.s3.com/thirdonsecond.mp4",
					},
				},
			},
			nil,
			true,
		},
		{
			"Empty fields",
			args{
				12,
				[]*shop.Media{
					{
						Id:    121,
						Label: "",
						Url:   "https://bucket.s3.com/firstonsecond.mp4",
					},
					{
						Id:    123,
						Label: "Third video on second article",
						Url:   "",
					},
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "TX error" {
				rt.Done()
			}

			if err := rt.updateVideos(tt.args.aid, tt.args.mds); (err != nil) != tt.wantErr {
				t.Errorf("requestTx.updateVideos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got, err := models.Videos(models.VideoWhere.ArticleID.EQ(tt.args.aid), qm.OrderBy(models.VideoColumns.Position)).All(rt.Ctx, rt.Tx)
				if len(tt.want) != len(got) {
					t.Errorf("requestTx.updateVideos() = %v, want %v", got, tt.want)
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				for k, w := range tt.want {
					if w.ID != got[k].ID || w.ArticleID != got[k].ArticleID || w.Position != got[k].Position || w.Label != got[k].Label || w.URL != got[k].URL {
						t.Errorf("requestTx.updateVideos() = \n%v\nwant\n%v", got[k], w)
					}
				}
				// check if we did not delete other videos
				img, err := models.FindVideo(rt.Ctx, rt.Tx, testVideos[0].ID, models.VideoColumns.ID)
				if err != nil || img.ID != testVideos[0].ID {
					t.Errorf("requestTx.updateVideos() = accidental delete")
				}
			}
		})
	}
}

var (
	testShopArts = []*shop.Article{
		{
			Id:          11,
			Created:     &timestamp.Timestamp{Seconds: 8000},
			Updated:     &timestamp.Timestamp{Seconds: 9000},
			Published:   false,
			Title:       "ID 11",
			Description: "This is the first article",
			Price:       "30000.99",
			Images: []*shop.Media{
				{
					Id:    111,
					Label: "First image on first article",
					Url:   "https://bucket.s3.com/firstonfirst.jpg",
				},
				{
					Id:    112,
					Label: "Second image on first article",
					Url:   "https://bucket.s3.com/secondonfirst.jpg",
				},
				{
					Id:    113,
					Label: "Third image on first article",
					Url:   "https://bucket.s3.com/thirdonfirst.jpg",
				},
			},
		},
		{
			Id:          12,
			Created:     &timestamp.Timestamp{Seconds: 1000},
			Updated:     &timestamp.Timestamp{Seconds: 2000},
			Published:   true,
			Title:       "ID 12",
			Description: "This is the second article",
			Price:       "12.12",
			Images: []*shop.Media{
				{
					Id:    121,
					Label: "First image on second article",
					Url:   "https://bucket.s3.com/firstonsecond.jpg",
				},
				{
					Id:    122,
					Label: "Second image on second article",
					Url:   "https://bucket.s3.com/secondonsecond.jpg",
				},
				{
					Id:    123,
					Label: "Third image on second article",
					Url:   "https://bucket.s3.com/thirdonsecond.jpg",
				},
			},
		},
		{
			Id:          13,
			Created:     &timestamp.Timestamp{Seconds: 3000},
			Updated:     &timestamp.Timestamp{Seconds: 4000},
			Published:   false,
			Title:       "ID 13",
			Description: "This is the third article",
			Price:       "22.99",
			Images: []*shop.Media{
				{
					Id:    131,
					Label: "First image on third article",
					Url:   "https://bucket.s3.com/firstonthird.jpg",
				},
				{
					Id:    132,
					Label: "Second image on third article",
					Url:   "https://bucket.s3.com/secondonthird.jpg",
				},
				{
					Id:    133,
					Label: "Third image on third article",
					Url:   "https://bucket.s3.com/thirdonthird.jpg",
				},
			},
			Promoted: true,
		},
	}

	testArticleView = &shop.Article{
		Id:          13,
		Created:     &timestamp.Timestamp{Seconds: 3000},
		Updated:     &timestamp.Timestamp{Seconds: 4000},
		Published:   false,
		Title:       "ID 13",
		Description: "This is the third article",
		Price:       "22.99",
		Images: []*shop.Media{
			{
				Id:    131,
				Label: "First image on third article",
				Url:   "https://bucket.s3.com/firstonthird.jpg",
			},
			{
				Id:    132,
				Label: "Second image on third article",
				Url:   "https://bucket.s3.com/secondonthird.jpg",
			},
			{
				Id:    133,
				Label: "Third image on third article",
				Url:   "https://bucket.s3.com/thirdonthird.jpg",
			},
		},
		Videos: []*shop.Media{
			{
				Id:    131,
				Label: "First video on third article",
				Url:   "https://bucket.s3.com/firstonthird.mp4",
			},
			{
				Id:    132,
				Label: "Second video on third article",
				Url:   "https://bucket.s3.com/secondonthird.mp4",
			},
			{
				Id:    133,
				Label: "Third video on third article",
				Url:   "https://bucket.s3.com/thirdonthird.mp4",
			},
		},
		Categories: []*shop.Category{
			{
				Id:    21,
				Label: "Full category",
			},
			{
				Id:    22,
				Label: "Only unpublished category",
			},
		},
		Baseprices: []*shop.BasePrice{
			{
				Id:    31,
				Label: "Cheap material",
				Price: "44.55",
			},
			{
				Id:    32,
				Label: "Premium material",
				Price: "99.99",
			},
		},
		Variants: []*shop.Variant{
			{
				Id:         41,
				Labels:     []string{"hello", "world"},
				Multiplier: "3.33",
			},
			{
				Id:         42,
				Labels:     []string{"foo", "bar"},
				Multiplier: "9.99",
			},
		},
		Promoted: true,
	}
)

func Test_requestTx_viewArticle(t *testing.T) {
	tests := []struct {
		name    string
		aid     int
		want    *shop.Article
		wantErr bool
	}{
		{
			"Existing article",
			13,
			testArticleView,
			false,
		},
		{
			"Non-existing article",
			89,
			nil,
			true,
		},
		{
			"DB Error",
			13,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", true)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.viewArticle(tt.aid)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.viewArticle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.viewArticle() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_listArticles(t *testing.T) {
	tests := []struct {
		name    string
		cond    *shop.ListConditions
		want    []*shop.Article
		wantErr bool
	}{
		{
			"DB Error",
			&shop.ListConditions{
				Fields: []shop.ArticleFields{shop.ArticleFields_ALL},
				Relations: &shop.ArticleRelations{
					Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
				},
			},
			nil,
			true,
		},
		{
			"All articles",
			&shop.ListConditions{
				Fields: []shop.ArticleFields{shop.ArticleFields_ALL},
				Relations: &shop.ArticleRelations{
					Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
				},
			},
			testShopArts,
			false,
		},
		{
			"Limit 1",
			&shop.ListConditions{
				Fields: []shop.ArticleFields{shop.ArticleFields_ALL},
				Relations: &shop.ArticleRelations{
					Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
				},
				Limits: &shop.Limits{
					Limit: 1,
				},
			},
			testShopArts[:1],
			false,
		},
		{
			"Limit & Offset 1",
			&shop.ListConditions{
				Fields: []shop.ArticleFields{shop.ArticleFields_ALL},
				Relations: &shop.ArticleRelations{
					Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
				},
				Limits: &shop.Limits{
					Limit:  1,
					Offset: 1,
				},
			},
			testShopArts[1:2],
			false,
		},
		{
			"Only published",
			&shop.ListConditions{
				OnlyPublished: true,
				Fields:        []shop.ArticleFields{shop.ArticleFields_ALL},
				Relations: &shop.ArticleRelations{
					Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
				},
			},
			testShopArts[1:2],
			false,
		},
		{
			"Only promoted",
			&shop.ListConditions{
				OnlyPromoted: true,
				Fields:       []shop.ArticleFields{shop.ArticleFields_ALL},
				Relations: &shop.ArticleRelations{
					Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
				},
			},
			testShopArts[2:3],
			false,
		},
		{
			"Only Category ID",
			&shop.ListConditions{
				OnlyCategoryId: 22,
				Fields:         []shop.ArticleFields{shop.ArticleFields_ALL},
				Relations: &shop.ArticleRelations{
					Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
				},
			},
			[]*shop.Article{
				testShopArts[0],
				testShopArts[2],
			},
			false,
		},
		{
			"Only Category Label",
			&shop.ListConditions{
				OnlyCategoryLabel: "Only unpublished category",
				Fields:            []shop.ArticleFields{shop.ArticleFields_ALL},
				Relations: &shop.ArticleRelations{
					Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
				},
			},
			[]*shop.Article{
				testShopArts[0],
				testShopArts[2],
			},
			false,
		},
		{
			"Field error",
			&shop.ListConditions{
				Fields: []shop.ArticleFields{
					shop.ArticleFields_CREATED,
					shop.ArticleFields_TITLE,
					shop.ArticleFields(99),
					shop.ArticleFields_PRICE,
				},
			},
			nil,
			true,
		},
		{
			"Only published with fields",
			&shop.ListConditions{
				OnlyPromoted: true,
				Fields: []shop.ArticleFields{
					shop.ArticleFields_TITLE,
					shop.ArticleFields_PRICE,
					shop.ArticleFields_PROMOTED,
					shop.ArticleFields_CREATED,
				},
				Relations: &shop.ArticleRelations{},
			},
			[]*shop.Article{{
				Created:  &timestamp.Timestamp{Seconds: 3000},
				Title:    "ID 13",
				Price:    "22.99",
				Promoted: true,
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", true)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}
			got, err := rt.listArticles(tt.cond)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.listArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Error("requestTx.listArticles() failed")
				for i, w := range tt.want {
					t.Logf("Got:\n%v\nWant:\n%v", got[i], w)
				}
			}
		})
	}
}

func Test_requestTx_listArticles_empty(t *testing.T) {
	migrateDown()
	migrations()

	defer func() {
		if err := testData(); err != nil {
			t.Fatal(err)
		}
	}()

	rt, err := tss.newTx(testCtx, "testing", true)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Done()

	cond := &shop.ListConditions{
		Fields: []shop.ArticleFields{shop.ArticleFields_ALL},
		Relations: &shop.ArticleRelations{
			Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
		},
	}

	got, err := rt.listArticles(cond)
	if err != nil {
		t.Errorf("requestTx.listArticles() error = %v, wantErr %v", err, false)
		return
	}

	want := []*shop.Article{}

	if !reflect.DeepEqual(got, want) {
		t.Error("requestTx.listArticles() failed")
		for i, w := range want {
			t.Logf("Got:\n%v\nWant:\n%v", got[i], w)
		}
	}
}

func Test_requestTx_checkDBErrors(t *testing.T) {
	type args struct {
		logMsg     string
		errs       []error
		warnNoRows bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			"Nil errs",
			args{
				"testing",
				nil,
				false,
			},
			nil,
		},
		{
			"Multi errors nill",
			args{
				"testing",
				[]error{nil, nil, nil},
				false,
			},
			nil,
		},
		{
			"No rows warning",
			args{
				"testing",
				[]error{nil, sql.ErrNoRows, nil},
				true,
			},
			status.Error(codes.NotFound, "testing"),
		},
		{
			"No rows ignored",
			args{
				"testing",
				[]error{nil, sql.ErrNoRows, nil},
				false,
			},
			nil,
		},
		{
			"Other error",
			args{
				"testing",
				[]error{nil, errors.New("other"), nil},
				true,
			},
			status.Error(codes.Internal, errDB),
		},
		{
			"Other error, followed by NoRows",
			args{
				"testing",
				[]error{nil, errors.New("other"), sql.ErrNoRows},
				true,
			},
			status.Error(codes.Internal, errDB),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", true)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()

			if err := rt.checkDBErrors(tt.args.logMsg, tt.args.errs, tt.args.warnNoRows); !errors.Is(err, tt.wantErr) {
				t.Errorf("requestTx.checkDBErrors() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_requestTx_deleteArticle(t *testing.T) {
	tests := []struct {
		name    string
		aid     int
		want    int64
		wantErr bool
	}{
		{
			"Existing article",
			13,
			13,
			false,
		},
		{
			"Non-Existing article",
			112,
			0,
			false,
		},
		{
			"DB Error",
			13,
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.deleteArticle(tt.aid)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.deleteArticle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("requestTx.deleteArticle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_shouldCalcPrice(t *testing.T) {
	tests := []struct {
		name    string
		art     *models.Article
		want    bool
		wantErr bool
	}{
		{
			"DB Error",
			testArticles[0],
			false,
			true,
		},
		{
			"Should not",
			testArticles[0],
			false,
			false,
		},
		{
			"Should",
			testArticles[2],
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.shouldCalcPrice(tt.art)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.shouldCalcPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("requestTx.shouldCalcPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_calcPrice(t *testing.T) {
	type args struct {
		art   *models.Article
		bpID  int
		vrtID int64
	}
	tests := []struct {
		name    string
		args    args
		want    *calculation
		wantErr bool
	}{
		{
			"Zero bpID",
			args{
				testArticles[2],
				0,
				41,
			},
			nil,
			true,
		},
		{
			"Zero vrtID",
			args{
				testArticles[2],
				31,
				0,
			},
			nil,
			true,
		},
		{
			"Non-existent base price",
			args{
				testArticles[2],
				29,
				41,
			},
			nil,
			true,
		},
		{
			"Calc 1",
			args{
				testArticles[2],
				31,
				41,
			},
			&calculation{
				decimal.New(1483515, 4),
				&shop.Details{
					BasePrice: &shop.BasePrice{
						Label: "Cheap material",
						Price: "44.55",
					},
					Variant: &shop.Variant{
						Labels:     []string{"hello", "world"},
						Multiplier: "3.33",
					},
				},
			},
			false,
		},
		{
			"Calc 2",
			args{
				testArticles[2],
				31,
				42,
			},
			&calculation{
				decimal.New(4450545, 4),
				&shop.Details{
					BasePrice: &shop.BasePrice{
						Label: "Cheap material",
						Price: "44.55",
					},
					Variant: &shop.Variant{
						Labels:     []string{"foo", "bar"},
						Multiplier: "9.99",
					},
				},
			},
			false,
		},
		{
			"Calc 3",
			args{
				testArticles[2],
				32,
				41,
			},
			&calculation{
				decimal.New(3329667, 4),
				&shop.Details{
					BasePrice: &shop.BasePrice{
						Label: "Premium material",
						Price: "99.99",
					},
					Variant: &shop.Variant{
						Labels:     []string{"hello", "world"},
						Multiplier: "3.33",
					},
				},
			},
			false,
		},
		{
			"Calc 4",
			args{
				testArticles[2],
				32,
				42,
			},
			&calculation{
				decimal.New(9989001, 4),
				&shop.Details{
					BasePrice: &shop.BasePrice{
						Label: "Premium material",
						Price: "99.99",
					},
					Variant: &shop.Variant{
						Labels:     []string{"foo", "bar"},
						Multiplier: "9.99",
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()

			got, err := rt.calcPrice(tt.args.art, tt.args.bpID, tt.args.vrtID)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.calcPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.calcPrice() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_newOrderArticle(t *testing.T) {
	js, err := json.Marshal(&shop.Details{
		BasePrice: &shop.BasePrice{
			Label: "Cheap material",
			Price: "44.55",
		},
		Variant: &shop.Variant{
			Labels:     []string{"hello", "world"},
			Multiplier: "3.33",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		so      *shop.Order_ArticleAmount
		want    *models.OrderArticle
		wantErr bool
	}{
		{
			"Article not found",
			&shop.Order_ArticleAmount{
				ArticleId: 99,
				Amount:    2,
			},
			nil,
			true,
		},
		{
			"Regular article",
			&shop.Order_ArticleAmount{
				ArticleId: 12,
				Amount:    2,
			},
			&models.OrderArticle{
				ArticleID: 12,
				Amount:    2,
				Title:     "ID 12",
				Price:     types.NewDecimal(decimal.New(1212, 2)),
			},
			false,
		},
		{
			"DB Error",
			&shop.Order_ArticleAmount{
				ArticleId: 12,
				Amount:    2,
			},
			nil,
			true,
		},
		{
			"Calc: unknown base price",
			&shop.Order_ArticleAmount{
				ArticleId:   13,
				Amount:      3,
				BasePriceId: 99,
				VariantId:   41,
			},
			nil,
			true,
		},
		{
			"Calc: success",
			&shop.Order_ArticleAmount{
				ArticleId:   13,
				Amount:      3,
				BasePriceId: 31,
				VariantId:   41,
			},
			&models.OrderArticle{
				ArticleID: 13,
				Amount:    3,
				Title:     "ID 13",
				Price:     types.NewDecimal(decimal.New(1483515, 4)),
				Details:   null.NewJSON(js, true),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.newOrderArticle(tt.so)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.newOrderArticle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.newOrderArticle() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

var testShopOrders = []*shop.Order{
	{
		FullName:      "Foo Bar",
		Email:         "foo@bar.com",
		Phone:         "0123456789",
		FullAddress:   "Office 1, No 7 Long street, Somewhere",
		Message:       "Gimme something",
		PaymentMethod: shop.Order_BANK_TRANSFER,
		Articles: []*shop.Order_ArticleAmount{
			{
				ArticleId: 11,
				Amount:    1,
			},
			{
				ArticleId: 12,
				Amount:    22,
			},
			{
				ArticleId:   13,
				Amount:      3,
				BasePriceId: 31,
				VariantId:   41,
			},
		},
	},
}

func Test_requestTx_newOrder(t *testing.T) {
	js, err := json.Marshal(&shop.Details{
		BasePrice: &shop.BasePrice{
			Label: "Cheap material",
			Price: "44.55",
		},
		Variant: &shop.Variant{
			Labels:     []string{"hello", "world"},
			Multiplier: "3.33",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	wantOrder := &models.Order{
		ID:            1,
		FullName:      "Foo Bar",
		Email:         "foo@bar.com",
		Phone:         "0123456789",
		FullAddress:   "Office 1, No 7 Long street, Somewhere",
		Message:       "Gimme something",
		PaymentMethod: "BANK_TRANSFER",
		Status:        "OPEN",
	}
	wantOrderArticles := []*models.OrderArticle{
		{
			OrderID:   1,
			ArticleID: 11,
			Amount:    1,
			ID:        1,
			Price:     types.NewDecimal(decimal.New(3000099, 2)),
			Title:     "ID 11",
			Details:   null.JSON{JSON: []byte{}},
		},
		{
			OrderID:   1,
			ArticleID: 12,
			Amount:    22,
			ID:        2,
			Price:     types.NewDecimal(decimal.New(1212, 2)),
			Title:     "ID 12",
			Details:   null.JSON{JSON: []byte{}},
		},
		{
			OrderID:   1,
			ArticleID: 13,
			Amount:    3,
			ID:        3,
			Price:     types.NewDecimal(decimal.New(1483515, 4)),
			Title:     "ID 13",
			Details:   null.NewJSON(js, true),
		},
	}

	tests := []struct {
		name     string
		so       *shop.Order
		want     *models.Order
		wantArts []*models.OrderArticle
		wantErr  error
	}{
		{
			"Order",
			testShopOrders[0],
			wantOrder,
			wantOrderArticles,
			nil,
		},
		{
			"DB Error",
			testShopOrders[0],
			nil,
			nil,
			status.Error(codes.Internal, errDB),
		},
		{
			"Empty order fields",
			&shop.Order{},
			nil,
			nil,
			status.Error(codes.InvalidArgument, "Missing required fields: Email, FullAddress, FullName, Phone"),
		},
		{
			"Empty order articles",
			&shop.Order{
				FullName:    "Foo Bar",
				Email:       "foo@bar.com",
				Phone:       "0123456789",
				FullAddress: "Office 1, No 7 Long street, Somewhere",
				Message:     "Gimme something",
			},
			nil,
			nil,
			status.Error(codes.InvalidArgument, errNoArts),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			order, err := rt.newOrder(tt.so)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("requestTx.newOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				got, err := models.FindOrder(rt.Ctx, rt.Tx, order.ID)
				if err != nil {
					t.Fatal(err)
				}
				goa, err := got.OrderArticles().All(rt.Ctx, rt.Tx)
				if err != nil {
					t.Fatal(err)
				}

				got.UpdatedAt, got.CreatedAt = time.Time{}, time.Time{}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("requestTx.newOrder() = \n%v\nwant\n%v", got, tt.want)
				}
				for k, w := range tt.wantArts {
					if w.Details.Valid != goa[k].Details.Valid {
						t.Errorf("requestTx.newOrder() Details.Valid = %v, want %v", goa[k].Details.Valid, w.Details.Valid)
					}
					if goa[k].Details.Valid {
						var wjs, gjs *shop.Details
						json.Unmarshal(w.Details.JSON, wjs)
						json.Unmarshal(goa[k].Details.JSON, gjs)
						if !reflect.DeepEqual(wjs, gjs) {
							t.Errorf("requestTx.newOrder() Details.JSON = %v, want %v", wjs, gjs)
						}
						w.Details = goa[k].Details
					}

					if !reflect.DeepEqual(w, goa[k]) {
						t.Errorf("requestTx.newOrder() = \n%v\nwant\n%v", goa[k], w)
						t.Log(string(goa[k].Details.JSON))
						t.Log(string(w.Details.JSON))
					}
				}
			}
		})
	}
	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}

var testShopOrderArticles = []*shop.Order_ArticleAmount{
	{
		ArticleId: 11,
		Amount:    1,
		Title:     "ID 11",
		Price:     "30000.99",
		Total:     "30000.99",
	},
	{
		ArticleId: 12,
		Amount:    3,
		Title:     "ID 12",
		Price:     "12.12",
		Total:     "36.36",
	},
	{
		ArticleId: 13,
		Amount:    5,
		Title:     "ID 13",
		Price:     "148.3515",
		Total:     "741.7575",
		Details: &shop.Details{
			BasePrice: &shop.BasePrice{
				Label: "Cheap material",
				Price: "44.55",
			},
			Variant: &shop.Variant{
				Labels:     []string{"hello", "world"},
				Multiplier: "3.33",
			},
		},
	},
}

func Test_requestTx_listOrders(t *testing.T) {
	tests := []struct {
		name    string
		cond    *shop.ListOrderConditions
		want    *shop.OrderList
		wantErr error
	}{
		{
			"DB Error",
			nil,
			nil,
			status.Error(codes.Internal, errDB),
		},
		{
			"With illigal status",
			nil,
			nil,
			status.Errorf(codes.Unimplemented, errEnum, "UNDEFINED"),
		},
		{
			"Status sent",
			&shop.ListOrderConditions{
				Status: shop.ListOrderConditions_SENT,
			},
			&shop.OrderList{List: []*shop.Order{{
				Id:            100,
				Created:       &timestamp.Timestamp{Seconds: 12},
				Updated:       &timestamp.Timestamp{Seconds: 34},
				FullName:      "What Is My Name",
				Email:         "me@example.com",
				Phone:         "0123456789",
				FullAddress:   "No. 7, Long Street, Somewhere",
				Message:       "My awesome order",
				PaymentMethod: shop.Order_CASH_ON_DELIVERY,
				Status:        shop.Order_SENT,
				Articles:      testShopOrderArticles,
				Sum:           "30779.1075",
			}}},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", true)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.listOrders(tt.cond)
			if !errors.Is(tt.wantErr, err) {
				t.Errorf("requestTx.listOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.listOrders() \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_getOrderArticles(t *testing.T) {
	tests := []struct {
		name    string
		order   *models.Order
		wantOaa []*shop.Order_ArticleAmount
		wantSum string
		wantErr bool
	}{
		{
			"DB Error",
			testOrders[0],
			nil,
			"",
			true,
		},
		{
			"Success",
			testOrders[0],
			testShopOrderArticles,
			"30779.1075",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", true)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			gotOaa, gotSum, err := rt.getOrderArticles(tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.getOrderArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOaa, tt.wantOaa) {
				t.Errorf("requestTx.getOrderArticles() = %v, want %v", gotOaa, tt.wantOaa)
			}
			if gotSum != tt.wantSum {
				t.Errorf("requestTx.getOrderArticles() = %v, want %v", gotSum, tt.wantSum)
			}
		})
	}
}

func Test_requestTx_sendOrderMail(t *testing.T) {
	type args struct {
		tmpl  string
		order *models.Order
	}
	tests := []struct {
		name string
		args
		wantErr bool
	}{
		{
			"Success",
			args{
				OrderMailTmpl,
				testOrders[0],
			},
			false,
		},
		{
			"Order error",
			args{
				OrderMailTmpl,
				testOrders[1],
			},
			true,
		},
		{
			"DB Error",
			args{
				OrderMailTmpl,
				testOrders[0],
			},
			true,
		},
		{
			"Mailer error",
			args{
				"foo",
				testOrders[0],
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", true)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			if err := rt.sendOrderMail(tt.args.tmpl, tt.args.order, rt.s.conf.Mail.ShopName); (err != nil) != tt.wantErr {
				t.Errorf("requestTx.sendOrderMail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_requestTx_saveOrder(t *testing.T) {
	want := *testOrders[1]
	want.Status = models.StatusOPEN
	want.Message = "Better status"

	tests := []struct {
		name    string
		so      *shop.Order
		want    *models.Order
		wantErr bool
	}{
		{
			"Adaptor error",
			nil,
			nil,
			true,
		},
		{
			"Success",
			&shop.Order{
				Id:            101,
				FullName:      "What Is My Name",
				Email:         "me@example.com",
				Phone:         "0123456789",
				FullAddress:   "No. 7, Long Street, Somewhere",
				Message:       "Better status",
				PaymentMethod: shop.Order_CASH_ON_DELIVERY,
				Status:        shop.Order_OPEN,
			},
			&want,
			false,
		},
		{
			"DB Error",
			&shop.Order{
				Id:            101,
				FullName:      "What Is My Name",
				Email:         "me@example.com",
				Phone:         "0123456789",
				FullAddress:   "No. 7, Long Street, Somewhere",
				Message:       "Better status",
				PaymentMethod: shop.Order_CASH_ON_DELIVERY,
				Status:        shop.Order_OPEN,
			},
			nil,
			true,
		},
		{
			"Unknown ID",
			&shop.Order{
				Id:            9999,
				FullName:      "What Is My Name",
				Email:         "me@example.com",
				Phone:         "0123456789",
				FullAddress:   "No. 7, Long Street, Somewhere",
				Message:       "Better status",
				PaymentMethod: shop.Order_CASH_ON_DELIVERY,
				Status:        shop.Order_OPEN,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.saveOrder(tt.so)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.saveOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.ID != tt.want.ID || got.FullName != tt.want.FullName || got.Email != tt.want.Email || got.Phone != tt.want.Phone || got.FullAddress != tt.want.FullAddress || got.Message != tt.want.Message || got.PaymentMethod != tt.want.PaymentMethod || got.Status != tt.want.Status {
					t.Errorf("requestTx.saveOrder() = \n%v\nwant\n%v", got, tt.want)
				}
			}
		})
	}
}

var testShopCategories = []*shop.Category{
	{
		Id:      20,
		Created: &timestamp.Timestamp{Seconds: 4000, Nanos: 0},
		Updated: &timestamp.Timestamp{Seconds: 5000, Nanos: 0},
		Label:   "Empty category",
	},
	{
		Id:      21,
		Created: &timestamp.Timestamp{Seconds: 6000, Nanos: 0},
		Updated: &timestamp.Timestamp{Seconds: 7000, Nanos: 0},
		Label:   "Full category",
	},
	{
		Id:      22,
		Created: &timestamp.Timestamp{Seconds: 8000, Nanos: 0},
		Updated: &timestamp.Timestamp{Seconds: 9000, Nanos: 0},
		Label:   "Only unpublished category",
	},
}

func Test_requestTx_saveCategories(t *testing.T) {
	tests := []struct {
		name    string
		list    []*shop.Category
		wantErr error
	}{
		{
			"Empty category",
			[]*shop.Category{{}},
			status.Errorf(codes.InvalidArgument, errMissing, "Label"),
		},
		{
			"Success",
			testShopCategories,
			nil,
		},
		{
			"DB Error",
			testShopCategories,
			status.Error(codes.Internal, errDB),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			if err := rt.saveCategories(tt.list); !errors.Is(err, tt.wantErr) {
				t.Errorf("requestTx.saveCategories() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func insertECatTx(rt *requestTx) (*models.Category, error) {
	ecat := &models.Category{
		ID:        909,
		CreatedAt: time.Unix(-62135596801, 0),
		Label:     "Invalid time",
		Position:  9,
	}
	return ecat, ecat.Insert(rt.Ctx, rt.Tx, boil.Infer())
}

func Test_requestTx_listCategories(t *testing.T) {
	tests := []struct {
		name    string
		clc     *shop.CategoryListConditions
		want    []*shop.Category
		wantErr bool
	}{
		{
			"All categories",
			nil,
			testShopCategories,
			false,
		},
		{
			"DB Error",
			nil,
			nil,
			true,
		},
		{
			"Adaptor error",
			nil,
			nil,
			true,
		},
		{
			"Only published",
			&shop.CategoryListConditions{OnlyPublishedArticles: true},
			testShopCategories[1:2],
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			if tt.name == "Adaptor error" {
				if _, err = insertECatTx(rt); err != nil {
					t.Fatal(err)
				}
			}

			got, err := rt.listCategories(tt.clc)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.listCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.listCategories() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_searchArticles(t *testing.T) {
	type args struct {
		ts *shop.TextSearch
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.ArticleList
		wantErr bool
	}{
		{
			name:    "Search success",
			args:    args{&shop.TextSearch{Text: "first"}},
			wantErr: false,
			want:    &shop.ArticleList{List: testShopArts[:1]},
		},
		{
			name:    "Search success 2",
			args:    args{&shop.TextSearch{Text: "article first"}},
			wantErr: false,
			want:    &shop.ArticleList{List: testShopArts[:1]},
		},
		{
			name:    "Search error",
			args:    args{&shop.TextSearch{Text: "first"}},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "searchArticles test", false)
			if err != nil {
				t.Error(err.Error())
				t.FailNow()
			}
			defer rt.Done()
			if tt.name == "Search error" {
				rt.Done()
			}
			got, err := rt.searchArticles(tt.args.ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.searchArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.searchArticles() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_suggest(t *testing.T) {
	firstArticle := &shop.Article{
		Id:          11,
		Created:     &timestamp.Timestamp{Seconds: 8000},
		Updated:     &timestamp.Timestamp{Seconds: 9000},
		Published:   false,
		Title:       "ID 11",
		Description: "This is the first article",
		Price:       "30000.99",
		Promoted:    false,
	}
	type args struct {
		ts *shop.TextSearch
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.SuggestionList
		wantErr bool
	}{
		{
			name:    "Suggest success",
			args:    args{&shop.TextSearch{Text: "firs:*"}},
			wantErr: false,
			want:    &shop.SuggestionList{Article: []*shop.Article{firstArticle}, Category: []*shop.Category{}},
		},
		{
			name:    "Suggest error",
			args:    args{&shop.TextSearch{Text: "fir:*"}},
			wantErr: true,
			want:    nil,
		},
		{
			name:    "Suggest error 2",
			args:    args{&shop.TextSearch{Text: ""}},
			wantErr: true,
			want:    nil,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, fmt.Sprintf("suggest test #%d", k), false)
			if tt.name == "Suggest error" {
				rt.Done()
			} else {
				defer rt.Done()
			}
			if err != nil {
				t.Error(err.Error())
				t.FailNow()
			}
			got, err := rt.suggest(tt.args.ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.suggest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.suggest() = \n%+v\n, want \n%+v\n", got, tt.want)
			}
		})
	}
}

func Test_requestTx_upsertBasePrice(t *testing.T) {
	tests := []struct {
		name    string
		sbp     *shop.BasePrice
		want    *shop.BasePrice
		wantErr bool
	}{
		{
			"Invalid price",
			&shop.BasePrice{
				Id:    99,
				Label: "Foo",
				Price: "Bar",
			},
			nil,
			true,
		},
		{
			"New BasePrice",
			&shop.BasePrice{
				Label: "New BasePrice",
				Price: "1.23",
			},
			&shop.BasePrice{
				Id:    1,
				Label: "New BasePrice",
				Price: "1.23",
			},
			false,
		},
		{
			"Update BasePrice",
			&shop.BasePrice{
				Id:    31,
				Label: "It got better",
				Price: "55.44",
			},
			&shop.BasePrice{
				Id:    31,
				Label: "It got better",
				Price: "55.44",
			},
			false,
		},
		{
			"DB Error",
			&shop.BasePrice{
				Id:    31,
				Label: "It got better",
				Price: "55.44",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.upsertBasePrice(tt.sbp)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.upsertBasePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				got.Created, got.Updated = nil, nil
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.upsertBasePrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_deleteBasePrice(t *testing.T) {
	tests := []struct {
		name    string
		sbp     *shop.BasePrice
		want    *shop.Deleted
		wantErr bool
	}{
		{
			"Missing Id",
			&shop.BasePrice{Id: 0},
			nil,
			true,
		},
		{
			"DB Error",
			&shop.BasePrice{Id: 31},
			nil,
			true,
		},
		{
			"Success",
			&shop.BasePrice{Id: 31},
			&shop.Deleted{Rows: 2},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.deleteBasePrice(tt.sbp)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.deleteBasePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.deleteBasePrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_listBasePrices(t *testing.T) {
	tests := []struct {
		name    string
		bpc     *shop.BasePriceListCondtions
		want    []*shop.BasePrice
		wantErr bool
	}{
		{
			"DB Error",
			nil,
			nil,
			true,
		},
		{
			"All",
			nil,
			testShopBasePrices,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.listBasePrices(tt.bpc)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.listBasePrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("requestTx.listBasePrices() =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_requestTx_encryptOrder(t *testing.T) {
	type args struct {
		order *models.Order
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test #1", args: args{&models.Order{ID: 2}}},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mobilpay.PublicCer = "config/sandbox.LK1F-GMV1-YWRD-7J6T-QD55.public.cer"
			mobilpay.PrivateKeyFile = "config/sandbox.LK1F-GMV1-YWRD-7J6T-QD55private.key"
			mp := mobilpay.CB{}
			mp.ParseKeys()
			rt, e := tss.newTx(testCtx, fmt.Sprintf("Test_requestTx_encryptOrder test #%d", k), false)
			defer rt.Done()
			if e != nil {
				log.Println(e.Error())
			}
			got, got1, err := rt.encryptOrder(tt.args.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.encryptOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == "" || got1 == "" && !tt.wantErr {
				t.Fail()
			}
		})
	}
}

func Test_requestTx_newMessage(t *testing.T) {
	tests := []struct {
		name    string
		sm      *shop.Message
		want    *models.Message
		wantErr bool
	}{
		{
			"Adaptor error",
			nil,
			nil,
			true,
		},
		{
			"Success",
			&shop.Message{
				Name:    "Muhlemmer",
				Email:   "foo@bar.com",
				Phone:   "+407889924345",
				Subject: "Hello world!",
				Message: "Spanac!",
			},
			&models.Message{
				ID:      1,
				Name:    "Muhlemmer",
				Email:   "foo@bar.com",
				Phone:   "+407889924345",
				Subject: "Hello world!",
				Message: "Spanac!",
			},
			false,
		},
		{
			"DB Error",
			&shop.Message{
				Name:    "Muhlemmer",
				Email:   "foo@bar.com",
				Phone:   "+407889924345",
				Subject: "Hello world!",
				Message: "Spanac!",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := tss.newTx(testCtx, "testing", false)
			if err != nil {
				t.Fatal(err)
			}
			defer rt.Done()
			if tt.name == "DB Error" {
				rt.Done()
			}

			got, err := rt.newMessage(tt.sm)
			if (err != nil) != tt.wantErr {
				t.Errorf("requestTx.newMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil {
				got.CreatedAt = tt.want.CreatedAt
				got.UpdatedAt = tt.want.UpdatedAt
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("requestTx.newMessage() =\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}

func Test_requestTx_sendMail(t *testing.T) {
	type args struct {
		tmpl string
		msg  *models.Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Success",
			args{
				MessageMailTmpl,
				&models.Message{
					ID:      1,
					Name:    "Muhlemmer",
					Email:   "foo@bar.com",
					Phone:   "+407889924345",
					Subject: "Hello world!",
					Message: "Spanac!",
				},
			},
			false,
		},
		{
			"Mailer error",
			args{
				"foo",
				&models.Message{
					ID:      1,
					Name:    "Muhlemmer",
					Email:   "foo@bar.com",
					Phone:   "+407889924345",
					Subject: "Hello world!",
					Message: "Spanac!",
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		rt, err := tss.newTx(testCtx, "testing", false)
		if err != nil {
			t.Fatal(err)
		}
		defer rt.Done()

		t.Run(tt.name, func(t *testing.T) {
			if err := rt.sendMail(tt.args.tmpl, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("requestTx.sendMail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
