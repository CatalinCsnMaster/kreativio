-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

alter table shop.articles add column search_index tsvector;

create index search_index_id on shop.articles using GIN (search_index);

create trigger search_index_update 
before insert or update
on shop.articles 
for each row execute procedure
tsvector_update_trigger(search_index,'pg_catalog.romanian',title,description);

-- +migrate Down

alter table shop.articles drop column search_index;

drop index if exists shop.search_index_id;

drop trigger search_index_update on shop.articles;