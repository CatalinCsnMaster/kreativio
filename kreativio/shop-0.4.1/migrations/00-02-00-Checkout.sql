-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

create type shop.status as enum
    ('UNDEFINED', 'OPEN', 'SENT', 'COMPLETED');

create type shop.payment as enum
    ('CASH_ON_DELIVERY', 'BANK_TRANSFER', 'ONLINE');

create table shop.orders (
    id serial not null primary key,
    created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
    full_name text not null,
    email text not null,
    phone text not null,
    full_address text not null,
    message text not null,
    payment_method shop.payment not null,
    status shop.status not null default 'OPEN'
);

create table shop.order_articles (
    order_id integer not null references shop.orders (id),
    article_id integer not null references shop.articles (id),
    amount integer not null,
    primary key(order_id, article_id)
);

-- +migrate Down

drop table shop.order_articles;
drop table shop.orders;
drop type shop.payment;
drop type shop.status;