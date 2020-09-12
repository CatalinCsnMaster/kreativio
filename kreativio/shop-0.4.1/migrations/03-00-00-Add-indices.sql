-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

CREATE INDEX published_articles_index ON shop.articles (published)
    WHERE published is true;

CREATE INDEX promoted_articles_index ON shop.articles (promoted)
    WHERE promoted is true;

CREATE INDEX category_label_index on shop.categories (label);

-- +migrate Down

DROP INDEX shop.category_label_index;
DROP INDEX shop.promoted_articles_index;
DROP INDEX shop.published_articles_index;