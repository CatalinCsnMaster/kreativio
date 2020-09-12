// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/moapis/shop"
	"github.com/moapis/shop/models"
	fj "github.com/valyala/fastjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_checkRequired(t *testing.T) {
	tests := []struct {
		name    string
		vals    map[string]interface{}
		wantErr error
	}{
		{
			"Nil",
			nil,
			nil,
		},
		{
			"All entries zero",
			map[string]interface{}{
				"Nil":    nil,
				"String": "",
				"Int":    0,
				"Float":  0.0,
			},
			status.Error(codes.InvalidArgument, "Missing required fields: Float, Int, Nil, String"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkRequired(tt.vals)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("checkRequired() error \n%v\nwantErr\n%v", err, tt.wantErr)
			}
		})
	}
}

func Test_articleMsgToModel(t *testing.T) {
	tests := []struct {
		name    string
		sa      *shop.Article
		want    *models.Article
		wantErr error
	}{
		{
			"Missing fields",
			&shop.Article{},
			nil,
			status.Error(codes.InvalidArgument, "Missing required fields: Description, Price, Title"),
		},
		{
			"All fields",
			&shop.Article{
				Id:          12345,
				Published:   true,
				Title:       "Ain't it good?",
				Description: "Best procuct to buy!",
				Price:       "20000.01",
				Promoted:    true,
			},
			&models.Article{
				ID:          12345,
				Published:   true,
				Title:       "Ain't it good?",
				Description: "Best procuct to buy!",
				Price:       types.NewDecimal(decimal.New(2000001, 2)), // 20000.01
				Promoted:    true,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := articleMsgToModel(tt.sa)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("articleMsgToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("articleMsgToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_variantsMsgToModel(t *testing.T) {
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
			"Missing fields",
			args{
				0,
				[]*shop.Variant{{}},
			},
			nil,
			status.Error(codes.InvalidArgument, "Missing required fields: ArticleID, Labels, Multiplier"),
		},
		{
			"All fields",
			args{
				999,
				[]*shop.Variant{
					{
						Id:         1001,
						Labels:     []string{"hello", "world"},
						Multiplier: "44.55",
					},
					{
						Id:         1002,
						Labels:     []string{"foo", "bar"},
						Multiplier: "99.99",
					},
				},
			},
			[]*models.Variant{
				{
					ID:         1001,
					ArticleID:  999,
					Labels:     []string{"hello", "world"},
					Multiplier: types.NewDecimal(decimal.New(4455, 2)),
				},
				{
					ID:         1002,
					ArticleID:  999,
					Labels:     []string{"foo", "bar"},
					Multiplier: types.NewDecimal(decimal.New(9999, 2)),
				},
			},
			nil,
		},
		{
			"Illigal multiplier",
			args{
				999,
				[]*shop.Variant{
					{
						Id:         1001,
						Labels:     []string{"hello", "world"},
						Multiplier: "spanac",
					},
				},
			},
			nil,
			status.Errorf(codes.InvalidArgument, errDecimal, varDecimal, "spanac"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := variantsMsgToModel(tt.args.aid, tt.args.sv)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("variantsMsgToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("variantsMsgToModel() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_imagesMsgToModel(t *testing.T) {
	type args struct {
		aid int
		sm  []*shop.Media
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.Image
		wantErr error
	}{
		{
			"Missing fields",
			args{
				0,
				[]*shop.Media{{}},
			},
			nil,
			status.Error(codes.InvalidArgument, "Missing required fields: ArticleID, Label, URL"),
		},
		{
			"All fields",
			args{
				999,
				[]*shop.Media{
					{
						Id:    1001,
						Label: "Thousand one",
						Url:   "https://example.com/image1001.jpg",
					},
					{
						Id:    1002,
						Label: "Thousand two",
						Url:   "https://example.com/image1002.jpg",
					},
				},
			},
			[]*models.Image{
				{
					ID:        1001,
					ArticleID: 999,
					Position:  1,
					Label:     "Thousand one",
					URL:       "https://example.com/image1001.jpg",
				},
				{
					ID:        1002,
					ArticleID: 999,
					Position:  2,
					Label:     "Thousand two",
					URL:       "https://example.com/image1002.jpg",
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := imagesMsgToModel(tt.args.aid, tt.args.sm)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("imagesMsgToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("imagesMsgToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_videosMsgToModel(t *testing.T) {
	type args struct {
		aid int
		sm  []*shop.Media
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.Video
		wantErr error
	}{
		{
			"Missing fields",
			args{
				0,
				[]*shop.Media{{}},
			},
			nil,
			status.Error(codes.InvalidArgument, "Missing required fields: ArticleID, Label, URL"),
		},
		{
			"All fields",
			args{
				999,
				[]*shop.Media{
					{
						Id:    1001,
						Label: "Thousand one",
						Url:   "https://example.com/video1001.jpg",
					},
					{
						Id:    1002,
						Label: "Thousand two",
						Url:   "https://example.com/video1002.jpg",
					},
				},
			},
			[]*models.Video{
				{
					ID:        1001,
					ArticleID: 999,
					Position:  1,
					Label:     "Thousand one",
					URL:       "https://example.com/video1001.jpg",
				},
				{
					ID:        1002,
					ArticleID: 999,
					Position:  2,
					Label:     "Thousand two",
					URL:       "https://example.com/video1002.jpg",
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := videosMsgToModel(tt.args.aid, tt.args.sm)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("videosMsgToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("videosMsgToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_timeModelToMsg(t *testing.T) {
	type args struct {
		created time.Time
		updated time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    *timestamp.Timestamp
		want1   *timestamp.Timestamp
		wantErr bool
	}{
		{
			"Valid times",
			args{
				time.Unix(12, 34),
				time.Unix(56, 78),
			},
			&timestamp.Timestamp{Seconds: 12, Nanos: 34},
			&timestamp.Timestamp{Seconds: 56, Nanos: 78},
			false,
		},
		{
			"Invalid created",
			args{
				time.Unix(-62135596801, 0),
				time.Unix(56, 78),
			},
			nil,
			nil,
			true,
		},
		{
			"Invalid updated",
			args{
				time.Unix(12, 34),
				time.Unix(-62135596801, 0),
			},
			nil,
			nil,
			true,
		},
		{
			"Zero times",
			args{},
			nil,
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := timeModelToMsg(tt.args.created, tt.args.updated)
			if (err != nil) != tt.wantErr {
				t.Errorf("timeModelToMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("timeModelToMsg() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("timeModelToMsg() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_articleModelToMsg(t *testing.T) {
	tests := []struct {
		name    string
		art     *models.Article
		want    *shop.Article
		wantErr bool
	}{
		{
			"Invalid time",
			&models.Article{
				ID:          9000,
				CreatedAt:   time.Unix(12, 34),
				UpdatedAt:   time.Unix(-62135596801, 0),
				Published:   true,
				Title:       "Hello world!",
				Description: "This is me saying hello.",
				Price:       types.NewDecimal(new(decimal.Big).SetFloat64(99.99)),
				Promoted:    true,
			},
			nil,
			true,
		},
		// valid cases covered by transaction_test
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := articleModelToMsg(tt.art)
			if (err != nil) != tt.wantErr {
				t.Errorf("articleModelToMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("articleModelToMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_orderMsgToModel(t *testing.T) {
	tests := []struct {
		name    string
		so      *shop.Order
		want    *models.Order
		wantErr error
	}{
		{
			"Missing fields",
			&shop.Order{},
			nil,
			status.Error(codes.InvalidArgument, "Missing required fields: Email, FullAddress, FullName, Phone"),
		},
		{
			"All fields",
			&shop.Order{
				FullName:      "Foo Bar",
				Email:         "foo@bar.com",
				Phone:         "0123456789",
				FullAddress:   "Office 1, No 7 Long street, Somewhere",
				Message:       "Gimme something",
				PaymentMethod: shop.Order_BANK_TRANSFER,
			},
			&models.Order{
				FullName:      "Foo Bar",
				Email:         "foo@bar.com",
				Phone:         "0123456789",
				FullAddress:   "Office 1, No 7 Long street, Somewhere",
				Message:       "Gimme something",
				PaymentMethod: models.PaymentBANK_TRANSFER,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orderMsgToModel(tt.so)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("orderMsgToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("orderMsgToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_orderArticlesMsgToModels(t *testing.T) {
	tests := []struct {
		name    string
		sa      []*shop.Order_ArticleAmount
		wantErr error
	}{
		{
			"Nil articles",
			nil,
			status.Error(codes.InvalidArgument, errNoArts),
		},
		{
			"Emtpy articles",
			[]*shop.Order_ArticleAmount{
				{ArticleId: 1, Amount: 0},
				{ArticleId: 2, Amount: 0},
				{ArticleId: 3, Amount: 0},
			},
			status.Error(codes.InvalidArgument, errNoArts),
		},
		{
			"Negative article",
			[]*shop.Order_ArticleAmount{
				{ArticleId: 1, Amount: 0},
				{ArticleId: 2, Amount: -1},
				{ArticleId: 3, Amount: 0},
			},
			status.Errorf(codes.InvalidArgument, errNegAmount, 2, -1),
		},
		{
			"Success",
			[]*shop.Order_ArticleAmount{
				{ArticleId: 1, Amount: 0},
				{ArticleId: 2, Amount: 10},
				{ArticleId: 3, Amount: 22},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkOrderArticles(tt.sa)
			if !errors.Is(tt.wantErr, err) {
				t.Errorf("checkOrderArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_orderArticlesModelsToMsg(t *testing.T) {
	tests := []struct {
		name    string
		arts    []*models.OrderArticle
		want    []*shop.Order_ArticleAmount
		want1   string
		wantErr bool
	}{
		{
			"nil",
			nil,
			[]*shop.Order_ArticleAmount{},
			"0",
			false,
		},
		{
			"Articles",
			[]*models.OrderArticle{
				{
					ArticleID: 44,
					Amount:    4,
					Title:     "Something 4",
					Price:     types.NewDecimal(decimal.New(4444, 2)), // 44.44
				},
				{
					ArticleID: 55,
					Amount:    5,
					Title:     "Something 5",
					Price:     types.NewDecimal(decimal.New(5555, 2)), // 55.55
				},
			},
			[]*shop.Order_ArticleAmount{
				{
					ArticleId: 44,
					Amount:    4,
					Title:     "Something 4",
					Price:     "44.44",
					Total:     "177.76",
				},
				{
					ArticleId: 55,
					Amount:    5,
					Title:     "Something 5",
					Price:     "55.55",
					Total:     "277.75",
				},
			},
			"455.51",
			false,
		},
		{
			"Unmarshal error",
			[]*models.OrderArticle{
				{
					ArticleID: 44,
					Amount:    4,
					Title:     "Something 4",
					Price:     types.NewDecimal(decimal.New(4444, 2)), // 44.44
					Details:   null.JSON{JSON: []byte("^"), Valid: true},
				},
			},
			nil,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := orderArticlesModelsToMsg(tt.arts)
			if (err != nil) != tt.wantErr {
				t.Errorf("orderArticlesModelsToMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("orderArticlesModelsToMsg() got = \n%v\nwant\n%v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("orderArticlesModelsToMsg() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_orderModelToMsg(t *testing.T) {
	valid := *testOrders[0]

	brokenTime := valid
	brokenTime.CreatedAt = time.Unix(-62135596801, 0)

	brokenPayment := valid
	brokenPayment.PaymentMethod = "FOOBAR"

	brokenStatus := valid
	brokenStatus.Status = "SPANAC"

	tests := []struct {
		name    string
		order   *models.Order
		want    *shop.Order
		wantErr error
	}{
		{
			"Success",
			&valid,
			&shop.Order{
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
			},
			nil,
		},
		{
			"Time conversion error",
			&brokenTime,
			nil,
			status.Error(codes.OutOfRange, "timestamp: seconds:-62135596801 before 0001-01-01"),
		},
		{
			"Payment method error",
			&brokenPayment,
			nil,
			status.Errorf(codes.Unimplemented, errEnum, "FOOBAR"),
		},
		{
			"Status error",
			&brokenStatus,
			nil,
			status.Errorf(codes.Unimplemented, errEnum, "SPANAC"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orderModelToMsg(tt.order)
			if !errors.Is(tt.wantErr, err) {
				t.Errorf("orderModelToMsg() error\n%v\nwantErr\n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("orderModelToMsg() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_orderUpdateMsgToModel(t *testing.T) {
	tests := []struct {
		name    string
		so      *shop.Order
		want    *models.Order
		wantErr error
	}{
		{
			"Missing fields",
			&shop.Order{},
			nil,
			status.Error(codes.InvalidArgument, "Missing required fields: ID"),
		},
		{
			"All fields",
			&shop.Order{
				Id:            22,
				FullName:      "Foo Bar",
				Email:         "foo@bar.com",
				Phone:         "0123456789",
				FullAddress:   "Office 1, No 7 Long street, Somewhere",
				Message:       "Gimme something",
				PaymentMethod: shop.Order_BANK_TRANSFER,
				Status:        shop.Order_COMPLETED,
			},
			&models.Order{
				ID:            22,
				FullName:      "Foo Bar",
				Email:         "foo@bar.com",
				Phone:         "0123456789",
				FullAddress:   "Office 1, No 7 Long street, Somewhere",
				Message:       "Gimme something",
				PaymentMethod: models.PaymentBANK_TRANSFER,
				Status:        models.StatusCOMPLETED,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orderUpdateMsgToModel(tt.so)
			if !errors.Is(tt.wantErr, err) {
				t.Errorf("orderModelToMsg() error\n%v\nwantErr\n%v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("orderUpdateMsgToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_categoryListMsgToModel(t *testing.T) {
	tests := []struct {
		name    string
		list    []*shop.Category
		want    []*models.Category
		wantErr error
	}{
		{
			"Empty category",
			[]*shop.Category{{}},
			nil,
			status.Errorf(codes.InvalidArgument, errMissing, "Label"),
		},
		{
			"Success",
			testShopCategories,
			[]*models.Category{
				{
					ID:       20,
					Label:    "Empty category",
					Position: 1,
				},
				{
					ID:       21,
					Label:    "Full category",
					Position: 2,
				},
				{
					ID:       22,
					Label:    "Only unpublished category",
					Position: 3,
				},
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := categoriesMsgToModel(tt.list)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("categoryListMsgToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("categoryListMsgToModel() = %v, want %v", got, tt.want)
			}
			for i, w := range tt.want {
				if !reflect.DeepEqual(got[i], w) {
					t.Errorf("categoryListMsgToModel() = \n%v\nwant\n%v", got[i], w)
				}
			}
		})
	}
}

func Test_categoriesModeltoMsg(t *testing.T) {
	tests := []struct {
		name    string
		cats    []*models.Category
		want    []*shop.Category
		wantErr bool
	}{
		{
			"Invalid time",
			[]*models.Category{
				{
					ID:        345,
					UpdatedAt: time.Unix(-62135596801, 0),
				},
			},
			nil,
			true,
		},
		{
			"Success",
			testCategories,
			testShopCategories,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := categoriesModeltoMsg(tt.cats)
			if (err != nil) != tt.wantErr {
				t.Errorf("categoriesModeltoMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("categoriesModeltoMsg() = %v, want %v", got, tt.want)
			}
			for i, w := range tt.want {
				if !reflect.DeepEqual(got[i], w) {
					t.Errorf("categoriesModeltoMsg() = \n%v\nwant\n%v", got[i], w)
				}
			}
		})
	}
}

func Test_timeBytesToMsg(t *testing.T) {
	type args struct {
		ca []byte
		ua []byte
	}
	tests := []struct {
		name        string
		args        args
		wantCreated *timestamp.Timestamp
		wantUpdated *timestamp.Timestamp
		wantErr     bool
	}{
		{
			"Nil input",
			args{},
			nil,
			nil,
			false,
		},
		{
			"Success",
			args{
				ca: []byte("2020-02-01T09:51:55+02:00"),
				ua: []byte("2020-02-10T15:03:33+02:00"),
			},
			&timestamp.Timestamp{Seconds: 1580543515},
			&timestamp.Timestamp{Seconds: 1581339813},
			false,
		},
		{
			"Invalid ca",
			args{
				ca: []byte("~"),
				ua: []byte("2020-02-10T15:03:33+02:00"),
			},
			nil,
			nil,
			true,
		},
		{
			"Invalid ua",
			args{
				ca: []byte("2020-02-01T09:51:55+02:00"),
				ua: []byte("~"),
			},
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCreated, gotUpdated, err := timeBytesToMsg(tt.args.ca, tt.args.ua)
			if (err != nil) != tt.wantErr {
				t.Errorf("timeBytesToMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCreated, tt.wantCreated) {
				t.Errorf("timeBytesToMsg() gotCreated = %v, want %v", gotCreated, tt.wantCreated)
			}
			if !reflect.DeepEqual(gotUpdated, tt.wantUpdated) {
				t.Errorf("timeBytesToMsg() gotUpdated = %v, want %v", gotUpdated, tt.wantUpdated)
			}
		})
	}
}

func Test_mediaValuesToMsg(t *testing.T) {
	js := `[{"id" : 21, "url" : "https://ex.com/i1", "label" : "foo"}, {"id" : 22, "url" : "https://ex.com/i2", "label" : "bar"}]`
	v, err := fj.Parse(js)
	if err != nil {
		t.Fatal(err)
	}
	va, err := v.Array()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		mds  []*fj.Value
		want []*shop.Media
	}{
		{
			"nil",
			nil,
			nil,
		},
		{
			"images",
			va,
			[]*shop.Media{
				{
					Id:    21,
					Url:   "https://ex.com/i1",
					Label: "foo",
				},
				{
					Id:    22,
					Url:   "https://ex.com/i2",
					Label: "bar",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mediaValuesToMsg(tt.mds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mediaValuesToMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_CategoryValuesToMsg(t *testing.T) {
	js := `[{"id" : 21, "label" : "foo"}, {"id" : 22, "label" : "bar"}]`
	v, err := fj.Parse(js)
	if err != nil {
		t.Fatal(err)
	}
	va, err := v.Array()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		fv   []*fj.Value
		want []*shop.Category
	}{
		{
			"nil",
			nil,
			nil,
		},
		{
			"categories",
			va,
			[]*shop.Category{
				{
					Id:    21,
					Label: "foo",
				},
				{
					Id:    22,
					Label: "bar",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := categoryValuesToMsg(tt.fv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mediaValuesToMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_basePriceValuesToMsg(t *testing.T) {
	js := `[{"id" : 21, "label" : "foo", "price" : "12.34"}, {"id" : 22, "label" : "bar", "price" : "56.78"}]`
	v, err := fj.Parse(js)
	if err != nil {
		t.Fatal(err)
	}
	va, err := v.Array()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		fv   []*fj.Value
		want []*shop.BasePrice
	}{
		{
			"nil",
			nil,
			nil,
		},
		{
			"categories",
			va,
			[]*shop.BasePrice{
				{
					Id:    21,
					Label: "foo",
					Price: "12.34",
				},
				{
					Id:    22,
					Label: "bar",
					Price: "56.78",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := basePriceValuesToMsg(tt.fv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mediaValuesToMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_variantValuesToMsg(t *testing.T) {
	js := `[{"id" : 21, "labels" : ["foo", "bar"], "multiplier" : "2.2"}, {"id" : 22, "labels" : ["hello", "world"], "multiplier" : "3.3"}]`
	v, err := fj.Parse(js)
	if err != nil {
		t.Fatal(err)
	}
	va, err := v.Array()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		fv   []*fj.Value
		want []*shop.Variant
	}{
		{
			"nil",
			nil,
			nil,
		},
		{
			"categories",
			va,
			[]*shop.Variant{
				{
					Id:         21,
					Labels:     []string{"foo", "bar"},
					Multiplier: "2.2",
				},
				{
					Id:         22,
					Labels:     []string{"hello", "world"},
					Multiplier: "3.3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := variantValuesToMsg(tt.fv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mediaValuesToMsg() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_articleValueToMsg(t *testing.T) {
	js := `{"id" : 1, "created_at" : "2020-02-01T09:51:55+02:00", "updated_at" : "2020-02-10T15:03:33+02:00", "published" : true, "title" : "Some title", "description" : "Some description", "price" : "7993.60", "promoted" : true, "images" : [{"id" : 21, "url" : "https://ex.com/i1", "label" : "foo"}, {"id" : 22, "url" : "https://ex.com/i2", "label" : "bar"}]}`
	v, err := fj.Parse(js)
	if err != nil {
		t.Fatal(err)
	}

	its := `{"created_at" : "~"}`
	vi, err := fj.Parse(its)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		art     *fj.Value
		wantSa  *shop.Article
		wantErr bool
	}{
		{
			"Success",
			v,
			&shop.Article{
				Id:          1,
				Created:     &timestamp.Timestamp{Seconds: 1580543515},
				Updated:     &timestamp.Timestamp{Seconds: 1581339813},
				Published:   true,
				Title:       "Some title",
				Description: "Some description",
				Price:       "7993.60",
				Promoted:    true,
				Images: []*shop.Media{
					{
						Id:    21,
						Url:   "https://ex.com/i1",
						Label: "foo",
					},
					{
						Id:    22,
						Url:   "https://ex.com/i2",
						Label: "bar",
					},
				},
			},
			false,
		},
		{
			"Invalid created_at",
			vi,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSa, err := articleValueToMsg(tt.art)
			if (err != nil) != tt.wantErr {
				t.Errorf("articleValueToMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSa, tt.wantSa) {
				t.Errorf("articleValueToMsg() = \n%v\nwant\n%v", gotSa, tt.wantSa)
			}
		})
	}
}

func Test_jsonScanner_Scan(t *testing.T) {
	tests := []struct {
		name    string
		v       interface{}
		wantErr error
	}{
		{
			"Empty JSON",
			nil,
			errJSONEmpty,
		},
		{
			"Assert error",
			22,
			errJSONAssert,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &jsonScanner{p: &fj.Parser{}}
			if err := s.Scan(tt.v); !errors.Is(err, tt.wantErr) {
				t.Errorf("jsonScanner.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_jsonScanner_Scan_err(t *testing.T) {
	s := &jsonScanner{p: &fj.Parser{}}
	if err := s.Scan([]byte("~")); err == nil {
		t.Errorf("jsonScanner.Scan() error = %v, wantErr %v", err, true)
	}
}

func Test_jsonScanner_articlesJSONtoMsg(t *testing.T) {
	tests := []struct {
		name    string
		js      string
		want    []*shop.Article
		wantErr bool
	}{
		{
			"Not an array",
			`{"title" : "spanac"}`,
			nil,
			true,
		},
		{
			"Invalid created_at",
			`[{"created_at" : "~"}]`,
			nil,
			true,
		},
		{
			"Success",
			`[{"id" : 1, "created_at" : "2020-02-01T09:51:55+02:00", "updated_at" : "2020-02-10T15:03:33+02:00", "published" : true, "title" : "Some title", "description" : "Some description", "price" : "7993.60", "promoted" : true, "images" : [{"id" : 21, "url" : "https://ex.com/i1", "label" : "foo"}, {"id" : 22, "url" : "https://ex.com/i2", "label" : "bar"}]}]`,
			[]*shop.Article{
				{
					Id:          1,
					Created:     &timestamp.Timestamp{Seconds: 1580543515},
					Updated:     &timestamp.Timestamp{Seconds: 1581339813},
					Published:   true,
					Title:       "Some title",
					Description: "Some description",
					Price:       "7993.60",
					Promoted:    true,
					Images: []*shop.Media{
						{
							Id:    21,
							Url:   "https://ex.com/i1",
							Label: "foo",
						},
						{
							Id:    22,
							Url:   "https://ex.com/i2",
							Label: "bar",
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &jsonScanner{
				p: &fj.Parser{},
			}

			var err error

			s.Value, err = s.p.Parse(tt.js)
			if err != nil {
				t.Fatal(err)
			}

			got, err := s.articlesJSONtoMsg()
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonScanner.articlesJSONtoMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jsonScanner.articlesJSONtoMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_basePriceMsgToModel(t *testing.T) {
	tests := []struct {
		name    string
		sbp     *shop.BasePrice
		want    *models.BasePrice
		wantErr bool
	}{
		{
			"Empty fields error",
			nil,
			nil,
			true,
		},
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
			"Success",
			&shop.BasePrice{
				Id:    99,
				Label: "Foo",
				Price: "1.23",
			},
			&models.BasePrice{
				ID:    99,
				Label: "Foo",
				Price: types.NewDecimal(decimal.New(123, 2)),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := basePriceMsgToModel(tt.sbp)
			if (err != nil) != tt.wantErr {
				t.Errorf("basePriceMsgToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("basePriceMsgToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_basePricesModeltoMsg(t *testing.T) {
	tests := []struct {
		name    string
		bps     []*models.BasePrice
		want    []*shop.BasePrice
		wantErr bool
	}{
		{
			"Invalid time",
			[]*models.BasePrice{
				{
					ID:        345,
					UpdatedAt: time.Unix(-62135596801, 0),
				},
			},
			nil,
			true,
		},
		{
			"Success",
			testBasePrices,
			testShopBasePrices,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := basePricesModeltoMsg(tt.bps)
			if (err != nil) != tt.wantErr {
				t.Errorf("basePricesModeltoMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("basePricesModeltoMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_messageMsgToModel(t *testing.T) {
	tests := []struct {
		name    string
		sm      *shop.Message
		want    *models.Message
		wantErr bool
	}{
		{
			"Nil Message",
			nil,
			nil,
			true,
		},
		{
			"Valid Message",
			&shop.Message{
				Name:    "Muhlemmer",
				Email:   "foo@bar.com",
				Phone:   "+407889924345",
				Subject: "Hello world!",
				Message: "Spanac!",
			},
			&models.Message{
				Name:    "Muhlemmer",
				Email:   "foo@bar.com",
				Phone:   "+407889924345",
				Subject: "Hello world!",
				Message: "Spanac!",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := messageMsgToModel(tt.sm)
			if (err != nil) != tt.wantErr {
				t.Errorf("messageMsgToModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("messageMsgToModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
