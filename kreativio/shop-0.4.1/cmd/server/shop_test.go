package main

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/moapis/shop"
	"github.com/moapis/shop/models"
)

var testToken string

func Test_shopServer_SaveArticle(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		req *shop.Article
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.ArticleID
		wantErr bool
	}{
		{
			"Updated article",
			args{
				testCtx,
				&shop.Article{
					Id:          13,
					Published:   true,
					Title:       "ID 13",
					Description: "This is the third article",
					Price:       "22.99",
					Images: []*shop.Media{
						{
							Id:    1002,
							Label: "one image",
							Url:   "one.jpg",
						},
					},
					Videos: []*shop.Media{
						{
							Id:    1002,
							Label: "one video",
							Url:   "one.mp4",
						},
					},
					Token: testToken,
				},
			},
			&shop.ArticleID{Id: 13},
			false,
		},
		{
			"Expired context",
			args{
				ectx,
				&shop.Article{
					Title:       "A new article",
					Description: "This is a new article",
					Price:       "123.44",
					Images: []*shop.Media{
						{
							Id:    1003,
							Label: "one image",
							Url:   "one.jpg",
						},
					},
					Videos: []*shop.Media{
						{
							Id:    1003,
							Label: "one video",
							Url:   "one.mp4",
						},
					},
					Token: testToken,
				},
			},
			nil,
			true,
		},
		{
			"Invalid token",
			args{
				testCtx,
				&shop.Article{
					Title:       "A new article",
					Description: "This is a new article",
					Price:       "123.44",
					Images: []*shop.Media{
						{
							Id:    1004,
							Label: "one image",
							Url:   "one.jpg",
						},
					},
					Videos: []*shop.Media{
						{
							Id:    1004,
							Label: "one video",
							Url:   "one.mp4",
						},
					},
					Token: "foobar",
				},
			},
			nil,
			true,
		},
		{
			"Duplicate title",
			args{
				testCtx,
				&shop.Article{
					Id:          80,
					Title:       "ID 13",
					Description: "This is the third article",
					Price:       "22.99",
					Images: []*shop.Media{
						{
							Id:    1005,
							Label: "one image",
							Url:   "one.jpg",
						},
					},
					Videos: []*shop.Media{
						{
							Id:    1005,
							Label: "one video",
							Url:   "one.mp4",
						},
					},
					Token: testToken,
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Log(tt.name)
		got, err := tss.SaveArticle(tt.args.ctx, tt.args.req)
		if (err != nil) != tt.wantErr {
			t.Errorf("shopServer.SaveArticle() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("shopServer.SaveArticle() = %v, want %v", got, tt.want)
		}
	}
	// Restore testdata
	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}

func Test_shopServer_ViewArticle(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		req *shop.ArticleID
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.Article
		wantErr bool
	}{
		{
			"Context error",
			args{
				ectx,
				&shop.ArticleID{Id: 13},
			},
			nil,
			true,
		},
		{
			"Missing ID",
			args{
				testCtx,
				nil,
			},
			nil,
			true,
		},
		{
			"Existing article",
			args{
				testCtx,
				&shop.ArticleID{Id: 13},
			},
			testArticleView,
			false,
		},
		{
			"Non-existing article",
			args{
				testCtx,
				&shop.ArticleID{Id: 89},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.ViewArticle(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.ViewArticle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.ViewArticle() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_shopServer_ListArticles(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		req *shop.ListConditions
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.ArticleList
		wantErr bool
	}{
		{
			"Context error",
			args{
				ectx,
				&shop.ListConditions{},
			},
			nil,
			true,
		},
		{
			"All articles",
			args{
				testCtx,
				&shop.ListConditions{
					Fields: []shop.ArticleFields{shop.ArticleFields_ALL},
					Relations: &shop.ArticleRelations{
						Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
					},
				},
			},
			&shop.ArticleList{List: testShopArts},
			false,
		},
		{
			"Only category ID",
			args{
				testCtx,
				&shop.ListConditions{
					OnlyCategoryId: 22,
					Fields:         []shop.ArticleFields{shop.ArticleFields_ALL},
					Relations: &shop.ArticleRelations{
						Images: []shop.MediaFields{shop.MediaFields_MD_ALL},
					},
				},
			},
			&shop.ArticleList{List: []*shop.Article{
				testShopArts[0],
				testShopArts[2],
			}},
			false,
		},
		{
			"Only published with fields",
			args{
				testCtx,
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
			},
			&shop.ArticleList{List: []*shop.Article{{
				Created:  &timestamp.Timestamp{Seconds: 3000},
				Title:    "ID 13",
				Price:    "22.99",
				Promoted: true,
			}}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.ListArticles(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.ListArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Error("requestTx.listArticles() failed")
				for i, w := range tt.want.List {
					t.Logf("Got:\n%v\nWant:\n%v", got.List[i], w)
				}
			}
		})
	}
}

func Test_shopServer_DeleteArticle(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		req *shop.ArticleID
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.Deleted
		wantErr bool
	}{
		{
			"Context error",
			args{
				ectx,
				&shop.ArticleID{
					Id:    13,
					Token: testToken,
				},
			},
			nil,
			true,
		},
		{
			"Missing ID",
			args{
				testCtx,
				&shop.ArticleID{
					Id:    0,
					Token: testToken,
				},
			},
			nil,
			true,
		},
		{
			"Nil request",
			args{
				testCtx,
				nil,
			},
			nil,
			true,
		},
		{
			"Bad token",
			args{
				testCtx,
				&shop.ArticleID{
					Id:    13,
					Token: "foobar",
				},
			},
			nil,
			true,
		},
		{
			"Success",
			args{
				testCtx,
				&shop.ArticleID{
					Id:    13,
					Token: testToken,
				},
			},
			&shop.Deleted{Rows: 13},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.DeleteArticle(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.DeleteArticle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.DeleteArticle() = %v, want %v", got, tt.want)
			}
		})
	}
	// Restore testdata
	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}

func Test_shopServer_Checkout(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	cc := *testConfig
	cc.Mail.Host = "foobar"

	ets, err := cc.newShopServer()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx context.Context
		req *shop.Order
	}
	tests := []struct {
		name    string
		server  *shopServer
		args    args
		want    *shop.OrderID
		wantErr bool
	}{
		{
			"Context error",
			tss,
			args{
				ectx,
				&shop.Order{},
			},
			nil,
			true,
		},
		{
			"Order",
			tss,
			args{
				testCtx,
				testShopOrders[0],
			},
			&shop.OrderID{Id: 1},
			false,
		},
		{
			"Mail error",
			ets,
			args{
				testCtx,
				testShopOrders[0],
			},
			nil,
			true,
		},
	}
	for k, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.server.Checkout(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.Checkout() #%d error = %v, wantErr %v", k, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.Checkout() = %v, want %v", got, tt.want)
			}
		})
	}
	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}

func Test_shopServer_ListOrders(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		req *shop.ListOrderConditions
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.OrderList
		wantErr bool
	}{
		{
			"Context error",
			args{
				ectx,
				&shop.ListOrderConditions{
					Status: shop.ListOrderConditions_SENT,
					Token:  testToken,
				},
			},
			nil,
			true,
		},
		{
			"Success",
			args{
				testCtx,
				&shop.ListOrderConditions{
					Status: shop.ListOrderConditions_SENT,
					Token:  testToken,
				},
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
			false,
		},
		{
			"Bad token",
			args{
				testCtx,
				&shop.ListOrderConditions{
					Status: shop.ListOrderConditions_SENT,
					Token:  "fooBar",
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.ListOrders(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.ListOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.ListOrders() = \n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func Test_shopServer_SaveOrder(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	cc := *testConfig
	cc.Mail.Host = "foobar"

	ets, err := cc.newShopServer()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx context.Context
		req *shop.Order
	}
	tests := []struct {
		name    string
		server  *shopServer
		args    args
		want    *shop.OrderID
		wantErr bool
	}{
		{
			"Context error",
			tss,
			args{
				ectx,
				&shop.Order{
					Id:            101,
					FullName:      "What Is My Name",
					Email:         "me@example.com",
					Phone:         "0123456789",
					FullAddress:   "No. 7, Long Street, Somewhere",
					Message:       "Better status",
					PaymentMethod: shop.Order_CASH_ON_DELIVERY,
					Status:        shop.Order_OPEN,
					Token:         testToken,
				},
			},
			nil,
			true,
		},
		{
			"Success",
			tss,
			args{
				testCtx,
				&shop.Order{
					Id:            101,
					FullName:      "What Is My Name",
					Email:         "me@example.com",
					Phone:         "0123456789",
					FullAddress:   "No. 7, Long Street, Somewhere",
					Message:       "Better status",
					PaymentMethod: shop.Order_CASH_ON_DELIVERY,
					Status:        shop.Order_OPEN,
					Token:         testToken,
				},
			},
			&shop.OrderID{Id: 101},
			false,
		},
		{
			"Invalid token",
			tss,
			args{
				testCtx,
				&shop.Order{
					Id:            101,
					FullName:      "What Is My Name",
					Email:         "me@example.com",
					Phone:         "0123456789",
					FullAddress:   "No. 7, Long Street, Somewhere",
					Message:       "Better status",
					PaymentMethod: shop.Order_CASH_ON_DELIVERY,
					Status:        shop.Order_OPEN,
					Token:         "FOO",
				},
			},
			nil,
			true,
		},
		{
			"Unknown ID",
			tss,
			args{
				testCtx,
				&shop.Order{
					Id:            9999,
					FullName:      "What Is My Name",
					Email:         "me@example.com",
					Phone:         "0123456789",
					FullAddress:   "No. 7, Long Street, Somewhere",
					Message:       "Better status",
					PaymentMethod: shop.Order_CASH_ON_DELIVERY,
					Status:        shop.Order_OPEN,
					Token:         testToken,
				},
			},
			nil,
			true,
		},
		{
			"Mail error",
			ets,
			args{
				testCtx,
				&shop.Order{
					Id:            101,
					FullName:      "What Is My Name",
					Email:         "me@example.com",
					Phone:         "0123456789",
					FullAddress:   "No. 7, Long Street, Somewhere",
					Message:       "Better status",
					PaymentMethod: shop.Order_CASH_ON_DELIVERY,
					Status:        shop.Order_OPEN,
					Token:         testToken,
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.server.SaveOrder(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.SaveOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.SaveOrder() = %v, want %v", got, tt.want)
			}
		})
	}
	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}

func insertECat() (*models.Category, error) {
	rt, err := tss.newTx(testCtx, "insterECat", false)
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	ecat, err := insertECatTx(rt)
	if err != nil {
		return nil, err
	}
	return ecat, rt.Commit()
}

func deleteECat(ecat *models.Category) error {
	rt, err := tss.newTx(testCtx, "insterECat", false)
	if err != nil {
		return err
	}
	defer rt.Done()

	if _, err = ecat.Delete(rt.Ctx, rt.Tx); err != nil {
		return err
	}
	return rt.Commit()
}

func Test_shopServer_SaveCategories(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		req *shop.CategoryList
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.CategoryList
		wantErr bool
	}{
		{
			"Context error",
			args{
				ectx,
				&shop.CategoryList{
					List:  testShopCategories,
					Token: testToken,
				},
			},
			nil,
			true,
		},
		{
			"Success",
			args{
				testCtx,
				&shop.CategoryList{
					List:  testShopCategories,
					Token: testToken,
				},
			},
			&shop.CategoryList{List: testShopCategories},
			false,
		},
		{
			"Empty category",
			args{
				testCtx,
				&shop.CategoryList{
					List:  []*shop.Category{{}},
					Token: testToken,
				},
			},
			nil,
			true,
		},
		{
			"Adaptor error",
			args{
				testCtx,
				&shop.CategoryList{
					List:  testShopCategories,
					Token: testToken,
				},
			},
			nil,
			true,
		},
		{
			"Bad token",
			args{
				testCtx,
				&shop.CategoryList{
					List:  testShopCategories,
					Token: "fooBar",
				},
			},
			nil,
			true,
		},
		{
			"Shuffeled and new",
			args{
				testCtx,
				&shop.CategoryList{
					List: []*shop.Category{
						testShopCategories[1],
						testShopCategories[0],
						{Label: "A new cat"},
						testShopCategories[2],
					},
					Token: testToken,
				},
			},
			&shop.CategoryList{List: []*shop.Category{
				testShopCategories[1],
				testShopCategories[0],
				{
					Id:    1,
					Label: "A new cat",
				},
				testShopCategories[2],
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Adaptor error" {
				ecat, err := insertECat()
				if err != nil {
					t.Fatal(err)
				}
				defer deleteECat(ecat)
			}

			got, err := tss.SaveCategories(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.SaveCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got.GetList()) != len(tt.want.GetList()) {
				t.Fatalf("shopServer.SaveCategories() = %v, want %v", got, tt.want)
			}

			for i, w := range tt.want.GetList() {
				if w.GetId() != got.GetList()[i].GetId() || w.GetLabel() != got.GetList()[i].GetLabel() {
					t.Errorf("shopServer.SaveCategories() = \n%v\nwant\n%v", got.GetList()[i], w)
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

func Test_shopServer_ListCategories(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		req *shop.CategoryListConditions
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.CategoryList
		wantErr bool
	}{
		{
			"Context error",
			args{
				ectx,
				nil,
			},
			nil,
			true,
		},
		{
			"All categories",
			args{
				testCtx,
				&shop.CategoryListConditions{OnlyPublishedArticles: false},
			},
			&shop.CategoryList{List: testShopCategories},
			false,
		},
		{
			"Adaptor error",
			args{
				testCtx,
				&shop.CategoryListConditions{OnlyPublishedArticles: false},
			},
			nil,
			true,
		},
		{
			"Only published",
			args{
				testCtx,
				&shop.CategoryListConditions{OnlyPublishedArticles: true},
			},
			&shop.CategoryList{List: testShopCategories[1:2]},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Adaptor error" {
				ecat, err := insertECat()
				if err != nil {
					t.Fatal(err)
				}
				defer deleteECat(ecat)
			}

			got, err := tss.ListCategories(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.ListCategories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got.GetList()) != len(tt.want.GetList()) {
				t.Fatalf("shopServer.ListCategories() = %v, want %v", got, tt.want)
			}

			for i, w := range tt.want.GetList() {
				if w.GetId() != got.GetList()[i].GetId() || w.GetLabel() != got.GetList()[i].GetLabel() {
					t.Errorf("shopServer.ListCategories() = \n%v\nwant\n%v", got.GetList()[i], w)
				}
			}
		})
	}
}

func Test_shopServer_SearchArticles(t *testing.T) {
	ectx, cfn := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*3))
	cfn()
	type args struct {
		ctx context.Context
		req *shop.TextSearch
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.ArticleList
		wantErr bool
	}{
		{
			name:    "Search success",
			args:    args{ctx: testCtx, req: &shop.TextSearch{Text: "first"}},
			want:    &shop.ArticleList{List: testShopArts[:1]},
			wantErr: false,
		},
		{
			name:    "Search fail",
			args:    args{ctx: ectx, req: &shop.TextSearch{Text: "first"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Search fail 2",
			args:    args{ctx: testCtx, req: nil},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.SearchArticles(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.SearchArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.SearchArticles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shopServer_Suggest(t *testing.T) {
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
	ectx, cfn := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*3))
	cfn()
	type args struct {
		ctx context.Context
		req *shop.TextSearch
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.SuggestionList
		wantErr bool
	}{
		{
			name:    "Search success",
			args:    args{ctx: testCtx, req: &shop.TextSearch{Text: "first"}},
			want:    &shop.SuggestionList{Article: []*shop.Article{firstArticle}, Category: []*shop.Category{}},
			wantErr: false,
		},
		{
			name:    "Search fail",
			args:    args{ctx: ectx, req: &shop.TextSearch{Text: "first"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Search fail 2",
			args:    args{ctx: testCtx, req: nil},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.Suggest(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.SearchArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.SearchArticles() = \n%v\n, want \n%v\n", got, tt.want)
			}
		})
	}
}

func Test_shopServer_SaveBasePrice(t *testing.T) {
	type args struct {
		ctx context.Context
		sbp *shop.BasePrice
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.BasePrice
		wantErr bool
	}{
		{
			"Auth error",
			args{
				testCtx,
				&shop.BasePrice{
					Label: "New BasePrice",
					Price: "1.23",
					Token: "foo",
				},
			},
			nil,
			true,
		},
		{
			"New BasePrice",
			args{
				testCtx,
				&shop.BasePrice{
					Label: "New BasePrice",
					Price: "1.23",
					Token: testToken,
				},
			},
			&shop.BasePrice{
				Id:    1,
				Label: "New BasePrice",
				Price: "1.23",
			},
			false,
		},
		{
			"Invalid price",
			args{
				testCtx,
				&shop.BasePrice{
					Label: "New BasePrice",
					Price: "foo",
					Token: testToken,
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.SaveBasePrice(tt.args.ctx, tt.args.sbp)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.SaveBasePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil {
				got.Created, got.Updated = nil, nil
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.SaveBasePrice() = %v, want %v", got, tt.want)
			}
		})
	}

	// Restore test data
	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}

func Test_shopServer_DeleteBasePrice(t *testing.T) {
	type args struct {
		ctx context.Context
		sbp *shop.BasePrice
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.Deleted
		wantErr bool
	}{
		{
			"Auth error",
			args{
				testCtx,
				&shop.BasePrice{
					Id:    31,
					Token: "foo",
				},
			},
			nil,
			true,
		},
		{
			"Missing Id error",
			args{
				testCtx,
				&shop.BasePrice{
					Token: testToken,
				},
			},
			nil,
			true,
		},
		{
			"Success",
			args{
				testCtx,
				&shop.BasePrice{
					Id:    31,
					Token: testToken,
				},
			},
			&shop.Deleted{Rows: 2},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.DeleteBasePrice(tt.args.ctx, tt.args.sbp)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.DeleteBasePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.DeleteBasePrice() = %v, want %v", got, tt.want)
			}
		})
	}

	// Restore test data
	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}

func Test_shopServer_ListBasesPrices(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		bpc *shop.BasePriceListCondtions
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.BasePriceList
		wantErr bool
	}{
		{
			"Context error",
			args{
				ectx,
				nil,
			},
			nil,
			true,
		},
		{
			"All",
			args{
				testCtx,
				nil,
			},
			&shop.BasePriceList{List: testShopBasePrices},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.ListBasesPrices(tt.args.ctx, tt.args.bpc)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.ListBasesPrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.ListBasesPrices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_shopServer_SendMessage(t *testing.T) {
	ectx, cancel := context.WithCancel(context.Background())
	cancel()

	type args struct {
		ctx context.Context
		sm  *shop.Message
	}
	tests := []struct {
		name    string
		args    args
		want    *shop.MessageID
		wantErr bool
	}{
		{
			"Context error",
			args{
				ectx,
				nil,
			},
			nil,
			true,
		},
		{
			"Nil message error",
			args{
				testCtx,
				nil,
			},
			nil,
			true,
		},
		{
			"Success",
			args{
				testCtx,
				&shop.Message{
					Name:    "Muhlemmer",
					Email:   "foo@bar.com",
					Phone:   "+407889924345",
					Subject: "Hello world!",
					Message: "Spanac!",
				},
			},
			&shop.MessageID{Id: 1},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tss.SendMessage(tt.args.ctx, tt.args.sm)
			if (err != nil) != tt.wantErr {
				t.Errorf("shopServer.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shopServer.SendMessage() = %v, want %v", got, tt.want)
			}
		})
	}

	migrateDown()
	migrations()
	if err := testData(); err != nil {
		t.Fatal(err)
	}
}
