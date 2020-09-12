-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

CREATE OR REPLACE VIEW shop.order_articles_view
 AS
 SELECT oa.order_id,
    oa.article_id,
    oa.amount,
    a.title,
    a.price
   FROM shop.articles a
     JOIN shop.order_articles oa ON oa.article_id = a.id;


-- +migrate Down

DROP VIEW shop.order_articles_view;