-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

create table shop.messages(
    id serial primary key,
    created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null,
    name text not null,
    email text not null,
    phone text not null,
    subject text not null,
    message text not null
);

-- +migrate Down

drop table shop.messages;