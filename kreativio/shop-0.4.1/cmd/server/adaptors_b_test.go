// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

// +build adapt

package main

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/moapis/shop"
	"github.com/moapis/shop/models"
	fj "github.com/valyala/fastjson"
	"github.com/volatiletech/sqlboiler/v4/types"
)

func Benchmark_checkRequired(b *testing.B) {
	vals := map[string]interface{}{
		"one":   1,
		"two":   "two",
		"three": 3.3,
		"four":  4,
		"five":  "five",
		"six":   6.6,
	}
	for i := 0; i < b.N; i++ {
		checkRequired(vals)
	}
}

func Benchmark_articleMsgToModel(b *testing.B) {
	sa := &shop.Article{
		Id:          12345,
		Published:   true,
		Title:       "Ain't it good?",
		Description: "Best procuct to buy!",
		Price:       "20000.01",
		Promoted:    true,
	}
	for i := 0; i < b.N; i++ {
		articleMsgToModel(sa)
	}
}

func Benchmark_imagesMsgToModel(b *testing.B) {
	aid := 999
	sm := []*shop.Media{
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
	}
	for i := 0; i < b.N; i++ {
		imagesMsgToModel(aid, sm)
	}
}

func Benchmark_videosMsgToModel(b *testing.B) {
	aid := 999
	sm := []*shop.Media{
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
	}
	for i := 0; i < b.N; i++ {
		videosMsgToModel(aid, sm)
	}
}

func Benchmark_timeModelToMsg(b *testing.B) {
	updated := time.Unix(12, 34)
	created := time.Unix(56, 78)

	for i := 0; i < b.N; i++ {
		timeModelToMsg(created, updated)
	}
}

func Benchmark_orderMsgToModel(b *testing.B) {
	so := &shop.Order{
		FullName:      "Foo Bar",
		Email:         "foo@bar.com",
		Phone:         "0123456789",
		FullAddress:   "Office 1, No 7 Long street, Somewhere",
		Message:       "Gimme something",
		PaymentMethod: shop.Order_BANK_TRANSFER,
	}
	for i := 0; i < b.N; i++ {
		orderMsgToModel(so)
	}
}

func Benchmark_checkOrderArticles(b *testing.B) {
	soa := []*shop.Order_ArticleAmount{
		{ArticleId: 1, Amount: 0},
		{ArticleId: 2, Amount: 0},
		{ArticleId: 3, Amount: 0},
	}
	for i := 0; i < b.N; i++ {
		checkOrderArticles(soa)
	}
}

func Benchmark_orderArticlesModelsToMsg(b *testing.B) {
	arts := []*models.OrderArticle{
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
	}
	for i := 0; i < b.N; i++ {
		orderArticlesModelsToMsg(arts)
	}
}

func Benchmark_orderModelToMsg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		orderModelToMsg(testOrders[0])
	}
}

var (
	articlesJSON    []byte
	articlesScanner *jsonScanner
)

func init() {
	var err error
	articlesJSON, err = ioutil.ReadFile("../../result.json")
	if err != nil {
		log.WithError(err).Fatal("init")
	}
	v, err := fj.ParseBytes(articlesJSON)
	if err != nil {
		log.WithError(err).Fatal("init")
	}
	articlesScanner = &jsonScanner{Value: v}
}

func Benchmark_jsonScanner_Scan(b *testing.B) {
	var p fj.Parser
	for i := 0; i < b.N; i++ {
		s := &jsonScanner{p: &p}
		s.Scan(articlesJSON)
	}
}

func Benchmark_jsonScanner_articlesJSONtoMsg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		articlesScanner.articlesJSONtoMsg()
	}
}
