-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

ALTER TABLE shop.order_articles
    ADD COLUMN id serial NOT NULL;

ALTER TABLE shop.order_articles
    ADD COLUMN title text;

ALTER TABLE shop.order_articles
    ADD COLUMN price numeric;

UPDATE shop.order_articles
	SET title = a.title, price = a.price
	FROM shop.articles a
	WHERE article_id = a.id;

ALTER TABLE shop.order_articles
    ALTER COLUMN title SET NOT NULL;

ALTER TABLE shop.order_articles
    ALTER COLUMN price SET NOT NULL;

ALTER TABLE shop.order_articles
    DROP CONSTRAINT order_articles_pkey;

ALTER TABLE shop.order_articles
    ADD PRIMARY KEY (id);

ALTER TABLE shop.order_articles
    ADD UNIQUE (order_id, article_id);

DROP VIEW shop.order_articles_view;

ALTER TABLE shop.articles
    ALTER COLUMN price TYPE numeric;

-- +migrate Down

ALTER TABLE shop.articles
    ALTER COLUMN price TYPE double precision;

CREATE OR REPLACE VIEW shop.order_articles_view
 AS
 SELECT oa.order_id,
    oa.article_id,
    oa.amount,
    a.title,
    a.price
   FROM shop.articles a
     JOIN shop.order_articles oa ON oa.article_id = a.id;

ALTER TABLE shop.order_articles
    DROP CONSTRAINT order_articles_pkey;

ALTER TABLE shop.order_articles
    ADD PRIMARY KEY (order_id, article_id);

ALTER TABLE shop.order_articles
    DROP CONSTRAINT order_articles_order_id_article_id_key;

ALTER TABLE shop.order_articles DROP COLUMN price;
ALTER TABLE shop.order_articles DROP COLUMN title;
ALTER TABLE shop.order_articles DROP COLUMN id;
