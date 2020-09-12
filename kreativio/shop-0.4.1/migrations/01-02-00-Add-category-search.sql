-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

alter table shop.categories add column search_index tsvector;

create index search_cat_index_id on shop.categories using GIN (search_index);

create trigger search_cat_index_update 
before insert or update
on shop.categories 
for each row execute procedure
tsvector_update_trigger(search_index,'pg_catalog.romanian',label);

-- +migrate Down

alter table shop.categories drop column search_index;

drop index if exists shop.search_cat_index_id;

drop trigger search_cat_index_update on shop.categories;