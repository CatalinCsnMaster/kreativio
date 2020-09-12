-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

create schema shop;

create table shop.articles (
	id serial not null primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	published boolean not null default false,
	title text not null,
	description text not null,
	price double precision not null,
	unique(title)
);

create table shop.images (
	id serial not null primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	article_id integer not null references shop.articles (id),
	position integer not null,
	label text not null,
	url text not null,
	unique(article_id, position)
);

create table shop.videos (
	id serial not null primary key,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
	article_id integer not null references shop.articles (id),
	position integer not null,
	label text not null,
	url text not null,
	unique(article_id, position)
);

-- +migrate Down
drop table shop.videos;
drop table shop.images;
drop table shop.articles;
drop schema shop;
