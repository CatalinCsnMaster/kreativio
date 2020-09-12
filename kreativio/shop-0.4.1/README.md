[![Build Status](https://travis-ci.org/moapis/shop.svg?branch=master)](https://travis-ci.org/moapis/shop)
[![codecov](https://codecov.io/gh/moapis/shop/branch/master/graph/badge.svg)](https://codecov.io/gh/moapis/shop)
[![GoDoc](https://godoc.org/github.com/moapis/shop?status.svg)](https://godoc.org/github.com/moapis/shop)
[![Go Report Card](https://goreportcard.com/badge/github.com/moapis/shop)](https://goreportcard.com/report/github.com/moapis/shop)

# Shop API
Shop is an API for publishing, listing and viewing articles in a shop.

## Development

### Migrations

````
cd migrations
sql-migrate up
````

### Models

From project root, Re-generate the models with:

````
sqlboiler psql
go test ./models
````

#### Testdata

A test data set can be generated using the [testdata](https://github.com/moapis/testdata) tool.
It relies on wordlists, which are in a Git submodule of this repository.

After installing the tool and the wordlists, run:

````
testdata -schema testdata.json
````

Connection details and schema definitions are in the `testdata.json` file.
Please don't modify the Seed, Min and Max values for columns, as it changes the determinism of the data.

### Protocol buffers

The shop API server uses gRPC through protocol buffers generation. To regenerate the gRPC definitions, run:

````
protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative shop.proto
````

### Benchmarks

Currently benchmarks are split between adaptor code and data code. For the latter a test data set must be loaded before execution.

**Query builer**

````
$ go test -bench . ./builder
...
BenchmarkArticleListQueryAllAll-8          41234             26462 ns/op
BenchmarkArticleListQueryImgCats-8        124092              9852 ns/op
BenchmarkArticleListDefaults-8            168234              7596 ns/op
BenchmarkArticleListWithFilters-8         164244              8273 ns/op
````

**Adaptors**

````
$ go test -bench . -tags adapt ./cmd/server
...
Benchmark_checkRequired-8                        9526696               123 ns/op
Benchmark_articleMsgToModel-8                    1561299               755 ns/op
Benchmark_imagesMsgToModel-8                     1418565               810 ns/op
Benchmark_videosMsgToModel-8                     1517041               834 ns/op
Benchmark_timeModelToMsg-8                       8821671               136 ns/op
Benchmark_orderMsgToModel-8                      2246192               542 ns/op
Benchmark_checkOrderArticles-8                  15703666                75.1 ns/op
Benchmark_orderArticlesModelsToMsg-8              607324              1904 ns/op
Benchmark_orderModelToMsg-8                      4236904               290 ns/op
Benchmark_jsonScanner_Scan-8                       92625             12256 ns/op
Benchmark_jsonScanner_articlesJSONtoMsg-8          38359             31360 ns/op
````

**Article list** benchmarks are a work in progress. Currently it runs against a set of 10.000 articles, random-ish filtered on category ID, published and promoted. The query returns relations in JSON documents, which is than converted to the gRPC types.

````
$ go test -bench . -tags data -benchtime=1000x ./cmd/server
...
Benchmark_transaction_jsonQuery-8                   1000           3049734 ns/op
Benchmark_transaction_listArticles-8                1000           2964290 ns/op
Benchmark_shopServer_ListArticles_serial-8          1000           3120801 ns/op
Benchmark_shopServer_ListArticles_routines-8        1000           2996836 ns/op
````
