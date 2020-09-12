-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

create table shop.categories (
    id serial not null primary key,
    created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
    label text not null,
    position integer not null,
    unique(label)
);

create table shop.category_articles (
    category_id integer not null references shop.categories (id),
    article_id integer not null references shop.articles (id),
    primary key(category_id, article_id),
    unique(article_id, category_id)
);

-- +migrate Down

drop table shop.category_articles;
drop table shop.categories;