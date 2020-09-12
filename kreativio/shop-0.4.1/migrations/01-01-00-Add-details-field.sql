-- Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
-- Use of this source code is governed by a License that can be found in the LICENSE file.
-- SPDX-License-Identifier: BSD-3-Clause

-- +migrate Up

ALTER TABLE shop.order_articles
    ADD COLUMN details text[] null;

-- +migrate Down

ALTER TABLE shop.order_articles DROP COLUMN details;