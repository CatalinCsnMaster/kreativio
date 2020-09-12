// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

// +build data

package main

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	pg "github.com/moapis/multidb/drivers/postgresql"
	"github.com/moapis/shop"
	"github.com/moapis/shop/builder"
	fj "github.com/valyala/fastjson"
)

var (
	bss *shopServer
)

func init() {
	c := Default

	c.MultiDB.DBConf = &pg.Config{
		Nodes: []pg.Node{
			{
				Host: "localhost",
				Port: 5432,
			},
		},
		Params: pg.Params{
			DBname:          "shop",
			User:            "postgres",
			Password:        "",
			SSLmode:         "disable",
			Connect_timeout: 30,
		},
	}

	var err error
	if bss, err = c.newShopServer(); err != nil {
		log.WithError(err).Fatal("newShopServer")
	}

	for n := 0; n < 200; n++ {
		for i := int32(0); i <= 10; i++ {
			conds := []*shop.ListConditions{
				{
					OnlyCategoryId: i,
					Limits:         &shop.Limits{Offset: rand.Int31n(975)},
				},
				{
					OnlyCategoryId: i,
					OnlyPublished:  true,
					Limits:         &shop.Limits{Offset: rand.Int31n(475)},
				},
				{
					OnlyCategoryId: i,
					OnlyPublished:  true,
					OnlyPromoted:   true,
					Limits:         &shop.Limits{Offset: rand.Int31n(225)},
				},
			}

			benchListConditions = append(
				benchListConditions,
				conds...,
			)
			for _, c := range conds {
				q, args, err := builder.ArticleListQuery(c, "shop")
				if err != nil {
					log.WithError(err).Fatal("ArticleListQuery")
				}
				benchListQueries = append(
					benchListQueries,
					benchListQuery{q, args},
				)
			}
		}
	}

	rand.Shuffle(len(benchListConditions), func(x, y int) {
		benchListConditions[x], benchListConditions[y] = benchListConditions[y], benchListConditions[x]
	})

	rand.Shuffle(len(benchListQueries), func(x, y int) {
		benchListQueries[x], benchListQueries[y] = benchListQueries[y], benchListQueries[x]
	})
}

type benchListQuery struct {
	query string
	args  []interface{}
}

var (
	benchListConditions []*shop.ListConditions
	benchListQueries    []benchListQuery
)

func Benchmark_transaction_jsonQuery(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	rt, err := bss.newTx(ctx, "articleQuery", true)
	if err != nil {
		b.Fatal(err)
	}

	p := new(fj.Parser)

	for i := 0; i < b.N; i++ {
		js := &jsonScanner{p: p}
		rt.jsonQuery(
			js,
			benchListQueries[i].query,
			benchListQueries[i].args...,
		)
	}

	rt.Done()
	cancel()
}

func Benchmark_transaction_listArticles(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	rt, err := bss.newTx(ctx, "articleQuery", true)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		rt.listArticles(benchListConditions[i])
	}
	cancel()
}

func Benchmark_shopServer_ListArticles_serial(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	for i := 0; i < b.N; i++ {
		bss.ListArticles(ctx, benchListConditions[i])
	}
	cancel()
}

func Benchmark_shopServer_ListArticles_routines(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(b.N)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	for i := 0; i < b.N; i++ {
		go func(i int) {
			bss.ListArticles(ctx, benchListConditions[i])
			wg.Done()
		}(i)
	}
	wg.Wait()
	cancel()
}
