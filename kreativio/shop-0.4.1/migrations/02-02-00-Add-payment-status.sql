-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

create table shop.payment_status(
    id serial primary key,
    order_id integer not null,
    confirmation_xml xml not null,
    status text not null,
    created_at timestamptz default NOW()
);

-- +migrate Down

drop table shop.payment_status;