-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

alter table shop.articles 
    add column promoted boolean not null default false;

-- +migrate Down

alter table shop.articles
    drop column promoted;