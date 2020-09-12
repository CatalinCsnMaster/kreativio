// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/moapis/mailer"
	"github.com/moapis/shop"
	"github.com/moapis/shop/builder"
	"github.com/moapis/shop/mobilpay"
	"github.com/moapis/shop/models"
	"github.com/moapis/transaction"
	"github.com/sirupsen/logrus"
	fj "github.com/valyala/fastjson"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	errMissingToken = "JWT token missing"
	errFatal        = "Fatal I/O error"
	errDB           = "Database error"
	errTS           = "Timestamp conversion error"
	errNotFound     = "%s with %s %v not found"
	errVrtPrice     = "Article ID %d missing BasePrice or Variant ID"
)

// requestTx holds a transaction.Request and the shopServer
type requestTx struct {
	*transaction.Request
	s *shopServer
}

func (s *shopServer) newTx(ctx context.Context, method string, readOnly bool) (*requestTx, error) {
	rt, err := transaction.New(ctx, s.log.WithField("method", method), s.mdb, readOnly, s.conf.SQLRoutines)
	if err != nil {
		return nil, err
	}
	return &requestTx{rt, s}, nil
}

func (s *shopServer) newAuthTx(ctx context.Context, method string, readOnly bool, token string) (*requestTx, error) {
	entry := s.log.WithField("method", method)
	if token == "" {
		entry.Warn(errMissingToken)
		return nil, status.Error(codes.Unauthenticated, errMissingToken)
	}

	rt, err := s.tv.NewAuth(ctx, entry, s.mdb, readOnly, s.conf.SQLRoutines, token, s.conf.Groups[method]...)
	if err != nil {
		return nil, err
	}
	rt.Log = rt.Log.WithField("user", rt.Claims.Subject)
	rt.Log.Info("Authenticated request")

	return &requestTx{rt, s}, nil
}

func (rt *requestTx) upsertArticle(sa *shop.Article) (*models.Article, error) {
	art, err := articleMsgToModel(sa)
	if err != nil {
		rt.Log.WithError(err).Warn("articleMsgToModel")
	}

	idc := models.ArticleColumns.ID
	entry := rt.Log.WithField("article", art)
	if err := art.Upsert(rt.Ctx, rt.Tx, true, []string{idc}, boil.Blacklist(idc), boil.Infer()); err != nil {
		entry.WithError(err).Error("rt.upsertArticle")
		return nil, status.Error(codes.Internal, errDB)
	}
	entry.Debug("rt.upsertArticle")
	return art, nil
}

func (rt *requestTx) setArticleCategories(art *models.Article, sc []*shop.Category) error {
	entry := rt.Log.WithField("categories", sc)

	cats := make([]*models.Category, len(sc))
	for i, c := range sc {
		cats[i] = &models.Category{ID: int(c.GetId())}
	}
	if err := art.SetCategories(rt.Ctx, rt.Tx, false, cats...); err != nil {
		entry.WithError(err).Error("art.SetCategories")
		return status.Error(codes.Internal, errDB)
	}
	entry.Debug("art.SetCategories")

	return nil
}

func (rt *requestTx) setArticleBasePrices(art *models.Article, sbp []*shop.BasePrice) error {
	entry := rt.Log.WithField("basePrices", sbp)

	bases := make([]*models.BasePrice, len(sbp))
	for i, b := range sbp {
		bases[i] = &models.BasePrice{ID: int(b.GetId())}
		if bases[i].ID == 0 {
			entry.Errorf(errMissing, "ID")
			return status.Errorf(codes.InvalidArgument, errMissing, "ID")
		}
	}
	if err := art.SetBasePrices(rt.Ctx, rt.Tx, false, bases...); err != nil {
		entry.WithError(err).Error("art.SetBasePrices")
		return status.Error(codes.Internal, errDB)
	}
	entry.Debug("art.SetBasePrices")

	return nil
}

func (rt *requestTx) updateVariants(aid int, sv []*shop.Variant) error {
	vars, err := variantsMsgToModel(aid, sv)
	if err != nil {
		rt.Log.WithError(err).Warn("variantsMsgToModel")
		return err
	}

	// Cleanup old variants
	ra, err := models.Variants(models.VariantWhere.ArticleID.EQ(aid)).DeleteAll(rt.Ctx, rt.Tx)
	if err != nil {
		rt.Log.WithError(err).Error("updateVariants: cleanup")
		return status.Error(codes.Internal, errDB)
	}
	rt.Log.WithField("rows", ra).Debug("updateVariants: cleanup")

	for _, v := range vars {
		entry := rt.Log.WithField("variant", v)
		if err := v.Insert(rt.Ctx, rt.Tx, boil.Infer()); err != nil {
			entry.WithError(err).Error("rt.updateVariants")
			return status.Error(codes.Internal, errDB)
		}
		entry.Debug("rt.updateVariants")
	}
	return nil
}

func (rt *requestTx) updateImages(aid int, mds []*shop.Media) error {
	imgs, err := imagesMsgToModel(aid, mds)
	if err != nil {
		rt.Log.WithError(err).Warn("imagesMsgToModel")
		return err
	}

	// Cleanup old images
	ra, err := models.Images(models.ImageWhere.ArticleID.EQ(aid)).DeleteAll(rt.Ctx, rt.Tx)
	if err != nil {
		rt.Log.WithError(err).Error("updateImages: cleanup")
		return err
	}
	rt.Log.WithField("rows", ra).Debug("updateImages: cleanup")

	for _, img := range imgs {
		entry := rt.Log.WithField("image", img)
		if err := img.Insert(rt.Ctx, rt.Tx, boil.Infer()); err != nil {
			entry.WithError(err).Error("rt.updateImages")
			return status.Error(codes.Internal, errDB)
		}
		entry.Debug("rt.updateImages")
	}
	return nil
}

func (rt *requestTx) updateVideos(aid int, mds []*shop.Media) error {
	vids, err := videosMsgToModel(aid, mds)
	if err != nil {
		rt.Log.WithError(err).Warn("videosMsgToModel")
		return err
	}

	// Cleanup old videos
	ra, err := models.Videos(models.VideoWhere.ArticleID.EQ(aid)).DeleteAll(rt.Ctx, rt.Tx)
	if err != nil {
		rt.Log.WithError(err).Error("updateVideos: cleanup")
		return err
	}
	rt.Log.WithField("rows", ra).Debug("updateVideos: cleanup")

	for _, vid := range vids {
		idc := models.VideoColumns.ID
		entry := rt.Log.WithField("image", vid)
		if err := vid.Upsert(rt.Ctx, rt.Tx, true, []string{idc}, boil.Blacklist(idc), boil.Infer()); err != nil {
			entry.WithError(err).Error("rt.updateVideos")
			return status.Error(codes.Internal, errDB)
		}
		entry.Debug("rt.updateVideos")
	}
	return nil
}

func (rt *requestTx) viewArticle(aid int) (*shop.Article, error) {
	rt.Log = rt.Log.WithField("aid", aid)
	art, err := models.Articles(
		models.ArticleWhere.ID.EQ(aid),
		qm.Load(models.ArticleRels.Images),
		qm.Load(models.ArticleRels.Videos),
		qm.Load(models.ArticleRels.Categories),
		qm.Load(models.ArticleRels.BasePrices),
		qm.Load(models.ArticleRels.Variants),
	).One(rt.Ctx, rt.Tx)
	switch err {
	case nil:
	case sql.ErrNoRows:
		rt.Log.WithError(err).Warn("Not found")
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Article %d not found", aid))
	default:
		rt.Log.WithError(err).Error("models.Articles")
		return nil, status.Error(codes.Internal, errDB)
	}
	rt.Log = rt.Log.WithField("art", art)
	rt.Log.Debug("models.Articles")

	sa, err := articleModelToMsg(art)
	if err != nil {
		rt.Log.WithError(err).Error("articleModelToMsg")
		return nil, err
	}
	return sa, nil
}

var artJSPool fj.ParserPool

func (rt *requestTx) listArticles(cond *shop.ListConditions) ([]*shop.Article, error) {
	rt.Log = rt.Log.WithField("cond", cond)

	query, args, err := builder.ArticleListQuery(cond, "shop")
	if err != nil {
		rt.Log.WithError(err).Error("builder.ArticleListQuery")
		return nil, err
	}

	js := jsonScanner{
		p: artJSPool.Get(),
	}
	defer artJSPool.Put(js.p)

	err = rt.Tx.QueryRowContext(rt.Ctx, query, args...).Scan(&js)
	switch {
	case err == nil:

	// TODO: this is a temporary workaround for:
	// https://github.com/golang/go/issues/38099
	case strings.Contains(err.Error(), errJSONEmpty.Error()):
		rt.Log.WithError(err).Debug("jsonQuery")
		return []*shop.Article{}, nil
	default:
		rt.Log.WithError(err).Error("QueryRowContext")
		return nil, status.Error(codes.Internal, errDB)
	}

	list, err := js.articlesJSONtoMsg()
	if err != nil {
		rt.Log.WithError(err).WithField("json", js.String()).Error("articlesJSONtoMsg")
		return nil, err
	}

	return list, nil
}

func (rt *requestTx) checkDBErrors(logMsg string, errs []error, warnNoRows bool) error {
	for _, e := range errs {
		isNoRows := errors.Is(e, sql.ErrNoRows)

		switch {
		case e == nil:
		case warnNoRows && isNoRows: // warn requested, so emmit one
			rt.Log.WithError(e).Warn(logMsg)
			return status.Error(codes.NotFound, logMsg)
		case isNoRows: // no warn requested, error is ignored
		default: // Any other error should be raised as error
			rt.Log.WithError(e).Error(logMsg)
			return status.Error(codes.Internal, errDB)
		}
	}
	rt.Log.Debug(logMsg)
	return nil
}

const (
	delCategoryArticles  = "delete from shop.category_articles where article_id = $1;"
	delArticleBasePrices = "delete from shop.article_base_prices where article_id = $1;"
)

func (rt *requestTx) deleteArticle(aid int) (int64, error) {
	cat, err := queries.Raw(delCategoryArticles, aid).ExecContext(rt.Ctx, rt.Tx)
	if err != nil {
		rt.Log.WithError(err).Error("delCategoryArticles")
		return 0, status.Error(codes.Internal, errDB)
	}
	bp, err := queries.Raw(delArticleBasePrices, aid).ExecContext(rt.Ctx, rt.Tx)
	if err != nil {
		rt.Log.WithError(err).Error("delArticleBasePrices")
		return 0, status.Error(codes.Internal, errDB)
	}

	ra, errs := make([]int64, 6), make([]error, 6)
	ra[0], errs[0] = models.Videos(models.VideoWhere.ArticleID.EQ(aid)).DeleteAll(rt.Ctx, rt.Tx)
	ra[1], errs[1] = models.Images(models.ImageWhere.ArticleID.EQ(aid)).DeleteAll(rt.Ctx, rt.Tx)
	ra[2], errs[2] = models.Variants(models.VariantWhere.ArticleID.EQ(aid)).DeleteAll(rt.Ctx, rt.Tx)
	ra[3], errs[3] = models.Articles(models.ArticleWhere.ID.EQ(aid)).DeleteAll(rt.Ctx, rt.Tx)
	ra[4], errs[4] = cat.RowsAffected()
	ra[5], errs[5] = bp.RowsAffected()

	var total int64
	for _, n := range ra {
		total += n
	}
	rt.Log = rt.Log.WithFields(logrus.Fields{"ra_vid": ra[0], "ra_img": ra[1], "ra_art": ra[2], "total": total})

	return total, rt.checkDBErrors("deleteArticle", errs, false)
}

func (rt *requestTx) shouldCalcPrice(art *models.Article) (bool, error) {
	var bp, vt bool
	errs := make([]error, 2)
	bp, errs[0] = art.BasePrices().Exists(rt.Ctx, rt.Tx)
	vt, errs[1] = art.Variants().Exists(rt.Ctx, rt.Tx)

	return bp && vt, rt.checkDBErrors("shouldCalcPrice", errs, false)
}

type calculation struct {
	Price   *decimal.Big
	Details *shop.Details
}

func (rt *requestTx) calcPrice(art *models.Article, bpID int, vrtID int64) (*calculation, error) {
	entry := rt.Log.WithFields(logrus.Fields{"article": art, "bpID": bpID, "vrtID": vrtID})

	if bpID == 0 || vrtID == 0 {
		entry.Warnf(errVrtPrice, art.ID)
		return nil, status.Errorf(codes.InvalidArgument, errVrtPrice, art.ID)
	}

	var (
		bp  *models.BasePrice
		vrt *models.Variant
	)
	errs := make([]error, 2)
	bp, errs[0] = models.FindBasePrice(
		rt.Ctx, rt.Tx, bpID,
		models.BasePriceColumns.Price,
		models.BasePriceColumns.Label,
	)
	vrt, errs[1] = models.FindVariant(
		rt.Ctx, rt.Tx, vrtID,
		models.VariantColumns.Multiplier,
		models.VariantColumns.Labels,
	)

	if err := rt.checkDBErrors(
		fmt.Sprintf("BasePrice and/or Variant for Article ID %d", art.ID),
		errs, true,
	); err != nil {
		return nil, err
	}

	return &calculation{
		new(decimal.Big).Mul(bp.Price.Big, vrt.Multiplier.Big),
		&shop.Details{
			BasePrice: &shop.BasePrice{
				Label: bp.Label,
				Price: bp.Price.String(),
			},
			Variant: &shop.Variant{
				Labels:     vrt.Labels,
				Multiplier: vrt.Multiplier.String(),
			},
		},
	}, nil
}

func (rt *requestTx) newOrderArticle(so *shop.Order_ArticleAmount) (*models.OrderArticle, error) {
	aid := int(so.GetArticleId())
	entry := rt.Log.WithField("aid", aid)

	art, err := models.FindArticle(rt.Ctx, rt.Tx, aid)
	switch err {
	case nil:
		entry = entry.WithField("article", art)
	case sql.ErrNoRows:
		entry.WithError(err).Warn("newOrderArticle")
		return nil, status.Errorf(codes.NotFound, errNotFound, "Article", "ID", aid)
	default:
		entry.WithError(err).Error("newOrderArticle")
		return nil, status.Error(codes.Internal, errDB)
	}

	oa := &models.OrderArticle{
		ArticleID: aid,
		Amount:    int(so.GetAmount()),
		Title:     art.Title,
		Price:     art.Price, // Will be over-written if calculate price != nil
	}

	if should, err := rt.shouldCalcPrice(art); err != nil {
		return nil, err
	} else if !should {
		return oa, nil
	}

	calc, err := rt.calcPrice(art, int(so.GetBasePriceId()), so.GetVariantId())
	if err != nil {
		return nil, err
	}
	entry = entry.WithField("calc", calc)
	oa.Price = types.NewDecimal(calc.Price)

	js, err := json.Marshal(calc.Details)
	if err != nil {
		entry.WithError(err).Error("calc.Details Marshal")
		return nil, status.Error(codes.Internal, errFatal)
	}
	oa.Details = null.NewJSON(js, true)

	entry.Debug("newOrderArticle")
	return oa, nil
}

func (rt *requestTx) newOrder(so *shop.Order) (*models.Order, error) {
	order, err := orderMsgToModel(so)
	if err != nil {
		rt.Log.WithError(err).Warn("orderMsgToModel")
		return nil, err
	}

	sa := so.GetArticles()
	if err = checkOrderArticles(sa); err != nil {
		return nil, err
	}
	arts := make([]*models.OrderArticle, len(sa))

	for i, a := range sa {
		arts[i], err = rt.newOrderArticle(a)
		if err != nil {
			return nil, err
		}
	}

	errs := make([]error, 2)
	errs[0] = order.Insert(rt.Ctx, rt.Tx, boil.Infer())
	errs[1] = order.AddOrderArticles(rt.Ctx, rt.Tx, true, arts...)

	rt.Log = rt.Log.WithFields(logrus.Fields{"order": order, "order_id": order.ID})
	return order, rt.checkDBErrors("newOrder", errs, false)
}

func (rt *requestTx) getOrderArticles(order *models.Order) (soaa []*shop.Order_ArticleAmount, sum string, err error) {
	arts, err := order.OrderArticles().All(rt.Ctx, rt.Tx)
	if err != nil {
		return nil, "", status.Error(codes.Internal, errDB)
	}
	return orderArticlesModelsToMsg(arts)
}

func (*requestTx) orderListQms(cond *shop.ListOrderConditions) (qms []qm.QueryMod) {
	if cond.GetStatus() != shop.ListOrderConditions_ANY {
		qms = append(qms,
			models.OrderWhere.Status.EQ(
				cond.GetStatus().String(),
			),
		)
	}

	return qms
}

func (rt *requestTx) listOrders(cond *shop.ListOrderConditions) (*shop.OrderList, error) {
	orders, err := models.Orders(rt.orderListQms(cond)...).All(rt.Ctx, rt.Tx)
	if err != nil {
		rt.Log.WithError(err).Error("models.Orders")
		return nil, status.Error(codes.Internal, errDB)
	}
	list := make([]*shop.Order, len(orders))
	for i, o := range orders {
		if list[i], err = orderModelToMsg(o); err != nil {
			return nil, err
		}
		if list[i].Articles, list[i].Sum, err = rt.getOrderArticles(o); err != nil {
			return nil, err
		}
	}
	return &shop.OrderList{List: list}, nil
}

func (rt *requestTx) sendOrderMail(tmpl string, order *models.Order, subject string) error {
	msg, err := orderModelToMsg(order)
	if err != nil {
		return err
	}
	if msg.Articles, msg.Sum, err = rt.getOrderArticles(order); err != nil {
		return err
	}

	to := append(rt.s.conf.Mail.To, msg.GetEmail())

	headers := []mailer.Header{
		{Key: "from", Values: []string{rt.s.conf.Mail.From}},
		{Key: "subject", Values: []string{subject}},
		{Key: "to", Values: to},
	}

	data := struct {
		*shop.Order
		Created  time.Time
		Currency string
	}{
		msg,
		order.CreatedAt,
		rt.s.conf.Mail.Currency,
	}

	log := rt.Log.WithFields(logrus.Fields{"headers": headers, "data": data})

	if err := rt.s.mail.Send(headers, tmpl, data, to...); err != nil {
		log.WithError(err).Error("sendOrderMail")
		return status.Error(codes.Internal, "Mailer error")
	}
	log.Debug("sendOrderMail")
	return nil
}

func (rt *requestTx) saveOrder(so *shop.Order) (*models.Order, error) {
	order, err := orderUpdateMsgToModel(so)
	if err != nil {
		rt.Log.WithError(err).Warn("orderUpdateMsgToModel")
		return nil, err
	}
	rt.Log = rt.Log.WithField("order", order)

	if _, err = order.Update(rt.Ctx, rt.Tx, boil.Blacklist(models.OrderColumns.ID)); err != nil {
		rt.Log.WithError(err).Error("order.Update")
		return nil, status.Error(codes.Internal, errDB)
	}
	rt.Log.Debug("order.Update")

	err = order.Reload(rt.Ctx, rt.Tx)
	switch err {
	case nil:
		rt.Log.Debug("order.Reload")
		return order, nil
	case sql.ErrNoRows:
		rt.Log.WithError(err).Warn("order.Reload")
		return nil, status.Errorf(codes.NotFound, "Order ID %d Not Found", so.GetId())
	default:
		rt.Log.WithError(err).Error("order.Reload")
		return nil, status.Error(codes.Internal, errDB)
	}
}

func (rt *requestTx) saveCategories(list []*shop.Category) error {
	cats, err := categoriesMsgToModel(list)
	if err != nil {
		rt.Log.WithError(err).Warn("categoryListMsgToModel")
		return err
	}
	rt.Log = rt.Log.WithField("categories", cats)

	for _, c := range cats {
		entry := rt.Log.WithField("category", c)
		if err = c.Upsert(
			rt.Ctx, rt.Tx, true,
			[]string{models.CategoryColumns.ID},
			boil.Blacklist(
				models.CategoryColumns.ID,
				models.CategoryColumns.CreatedAt,
			),
			boil.Infer(),
		); err != nil {
			entry.WithError(err).Error("c.Upsert")
			return status.Error(codes.Internal, errDB)
		}
	}

	return nil
}

const (
	categoryArticlesJoin = "shop.category_articles ca on shop.categories.id = ca.category_id"
	articlesJoin         = "shop.articles a on ca.article_id = a.id"
)

var (
	categoryGroup = strings.Join([]string{models.TableNames.Categories, models.CategoryColumns.ID}, ".")
)

func (*requestTx) listCategoriesQms(clc *shop.CategoryListConditions) []qm.QueryMod {
	qms := []qm.QueryMod{
		qm.OrderBy(models.CategoryColumns.Position),
	}
	if clc.GetOnlyPublishedArticles() {
		qms = append(qms,
			qm.InnerJoin(categoryArticlesJoin),
			qm.InnerJoin(articlesJoin),
			qm.GroupBy(categoryGroup),
			qm.Where("a.published=?", true),
		)
	}

	return qms
}

func (rt *requestTx) listCategories(clc *shop.CategoryListConditions) ([]*shop.Category, error) {
	cats, err := models.Categories(rt.listCategoriesQms(clc)...).All(rt.Ctx, rt.Tx)
	if err != nil {
		rt.Log.WithError(err).Error("models.Categories.All")
		return nil, status.Error(codes.Internal, errDB)
	}
	rt.Log = rt.Log.WithField("categories", cats)
	rt.Log.Debug("models.Categories.All")

	scs, err := categoriesModeltoMsg(cats)
	if err != nil {
		rt.Log.WithError(err).Error("categoriesModeltoMsg")
		return nil, err
	}

	return scs, nil
}

const (
	searchWhere    = "search_index @@ plainto_tsquery(?, ?)"
	searchLanguage = "romanian"
	suggestWhere   = "search_index @@ to_tsquery(?, ?)"
)

func (rt *requestTx) searchArticles(ts *shop.TextSearch) (*shop.ArticleList, error) {
	if ts == nil || ts.Text == "" {
		rt.Log.Error("searchArticles invalid argument")
		return nil, status.Error(codes.InvalidArgument, "Invalid")
	}
	artQ, err := models.Articles(
		qm.Where(searchWhere, searchLanguage, strings.TrimSpace(ts.GetText())),
		qm.Load(models.ArticleRels.Images),
	).All(rt.Ctx, rt.Tx)
	if err != nil {
		rt.Log.WithError(err).Error("models.Articles")
		return nil, status.Error(codes.Internal, errDB)
	}
	list := &shop.ArticleList{List: make([]*shop.Article, len(artQ))}
	for k, v := range artQ {
		art, err := articleModelToMsg(v)
		if err != nil {
			rt.Log.WithError(err).Error("articleModelToMsg loop error")
			return nil, status.Error(codes.Internal, "articleModelToMsg")
		}
		list.List[k] = art
	}
	return list, nil
}

func (rt *requestTx) suggest(ts *shop.TextSearch) (*shop.SuggestionList, error) {
	var err [2]error
	var art models.ArticleSlice
	if ts == nil || ts.Text == "" {
		rt.Log.Error("suggest invalid argument")
		return nil, status.Error(codes.InvalidArgument, "Invalid")
	}
	art, err[0] = models.Articles(qm.Where(suggestWhere, searchLanguage, strings.TrimSpace(ts.GetText()))).All(rt.Ctx, rt.Tx)
	list := &shop.SuggestionList{
		Article: make([]*shop.Article, len(art)),
	}
	list.Category, err[1] = rt.suggestCategory(ts)
	for _, v := range err {
		if v != nil {
			rt.Log.WithError(v).Error("models")
			return nil, status.Error(codes.Internal, errDB)
		}
	}
	for k, v := range art {
		art, err := articleModelToMsg(v)
		if err != nil {
			rt.Log.WithError(err).Error("articleModelToMsg loop error")
			return nil, status.Error(codes.Internal, "articleModelToMsg")
		}
		list.Article[k] = art
	}
	return list, nil
}
func (rt *requestTx) suggestCategory(ts *shop.TextSearch) ([]*shop.Category, error) {
	cat, err := models.Categories(qm.Where(suggestWhere, searchLanguage, strings.TrimSpace(ts.GetText()))).All(rt.Ctx, rt.Tx)
	if err != nil {
		return nil, err
	}
	return categoriesModeltoMsg(cat)
}

func (rt *requestTx) upsertBasePrice(sbp *shop.BasePrice) (*shop.BasePrice, error) {
	bp, err := basePriceMsgToModel(sbp)
	if err != nil {
		return nil, err
	}
	idc := models.BasePriceColumns.ID
	if err := bp.Upsert(rt.Ctx, rt.Tx, true, []string{idc}, boil.Blacklist(idc), boil.Infer()); err != nil {
		return nil, status.Error(codes.Internal, errDB)
	}
	return basePriceModeltoMsg(bp)
}

const (
	delBasePriceArticles = "delete from shop.article_base_prices where base_price_id = $1;"
)

func (rt *requestTx) deleteBasePrice(sbp *shop.BasePrice) (*shop.Deleted, error) {
	bpid := int(sbp.GetId())
	if bpid == 0 {
		return nil, status.Errorf(codes.InvalidArgument, errMissing, "Id")
	}

	var (
		res  sql.Result
		del  shop.Deleted
		errs = make([]error, 3)
	)

	res, errs[0] = queries.Raw(delBasePriceArticles, bpid).ExecContext(rt.Ctx, rt.Tx)
	del.Rows, errs[1] = models.BasePrices(models.BasePriceWhere.ID.EQ(bpid)).DeleteAll(rt.Ctx, rt.Tx)

	if res != nil {
		var ra int64
		ra, errs[2] = res.RowsAffected()
		del.Rows += ra
	}

	if err := rt.checkDBErrors("deleteBasePrice", errs, false); err != nil {
		return nil, err
	}

	return &del, nil
}

func (rt *requestTx) listBasePrices(*shop.BasePriceListCondtions) ([]*shop.BasePrice, error) {
	bps, err := models.BasePrices().All(rt.Ctx, rt.Tx)
	if err != nil {
		return nil, status.Error(codes.Internal, errDB)
	}
	return basePricesModeltoMsg(bps)
}

func (rt *requestTx) encryptOrder(order *models.Order) (string, string, error) {
	err := make([]error, 3)
	var sum string
	var b []byte
	var encData, encKey string
	_, sum, err[0] = rt.getOrderArticles(order)
	orderXML := mobilpay.Request{}
	orderXML.Order.ID = strconv.Itoa(order.ID)
	orderXML.Order.Signature = mobilpay.Signature
	orderXML.Order.Type = "card"
	orderXML.Order.URL.Confirm = mobilpay.ConfirmURL
	orderXML.Order.URL.Return = mobilpay.ReturnURL
	orderXML.Order.Invoice.Amount = sum
	orderXML.Order.Invoice.Currency = "RON"
	orderXML.Order.Invoice.Details = "Order payment by Credit Card."
	nm := strings.Split(order.FullName, " ")
	orderXML.Order.Invoice.ContactInfo.Billing.Fname = strings.Join(nm[1:], " ")
	orderXML.Order.Invoice.ContactInfo.Billing.Lname = nm[0]
	orderXML.Order.Invoice.ContactInfo.Billing.Address = order.FullAddress
	orderXML.Order.Invoice.ContactInfo.Billing.MobilePhone = order.Phone
	orderXML.Order.Invoice.ContactInfo.Billing.Email = order.Email
	orderXML.Order.Invoice.ContactInfo.Shipping.SameAsBilling = "1"
	b, err[1] = xml.Marshal(&orderXML)
	encData, encKey, err[2] = mobilpay.Encrypt(mobilpay.Cert, b)
	for k, e := range err {
		if e != nil {
			rt.Log.WithError(e).Errorf("encryptOrder #%d error", k)
			return "", "", e
		}
	}
	return encData, encKey, nil
}

func (rt *requestTx) newMessage(sm *shop.Message) (*models.Message, error) {
	msg, err := messageMsgToModel(sm)
	if err != nil {
		return nil, err
	}

	if err = msg.Insert(rt.Ctx, rt.Tx, boil.Infer()); err != nil {
		return nil, status.Error(codes.Internal, errDB)
	}

	return msg, nil
}

func (rt *requestTx) sendMail(tmpl string, msg *models.Message) error {
	subject := fmt.Sprintf("%s: message #%d: %s", rt.s.conf.Mail.ShopName, msg.ID, msg.Subject)

	headers := []mailer.Header{
		{Key: "from", Values: []string{rt.s.conf.Mail.From}},
		{Key: "subject", Values: []string{subject}},
		{Key: "to", Values: rt.s.conf.Mail.To},
	}

	log := rt.Log.WithFields(logrus.Fields{"headers": headers, "msg": msg})

	if err := rt.s.mail.Send(headers, tmpl, msg, rt.s.conf.Mail.To...); err != nil {
		log.WithError(err).Error("sendMail")
		return status.Error(codes.Internal, "Mailer error")
	}
	log.Debug("sendOrderMail")
	return nil
}
