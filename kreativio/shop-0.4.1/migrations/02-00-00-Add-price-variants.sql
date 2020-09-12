-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

create table shop.base_prices (
    id serial not null primary key,
    created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
    label text not null,
    price numeric not null,
    unique(label)
);

create table shop.article_base_prices (
    article_id integer not null references shop.articles (id),
    base_price_id integer not null references shop.base_prices (id),
    primary key (article_id, base_price_id)
);

create table shop.variants (
    id bigserial not null primary key,
    created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
    article_id integer not null references shop.articles (id),
    labels text[] not null,
    multiplier numeric not null,
    unique(labels)
);

-- +migrate Down

drop table shop.variants;
drop table shop.article_base_prices;
drop table shop.base_prices;
