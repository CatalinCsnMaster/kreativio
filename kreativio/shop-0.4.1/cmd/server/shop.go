package main

import (
	"context"
	"fmt"

	"github.com/moapis/mailer"
	"github.com/moapis/multidb"
	"github.com/moapis/shop"
	"github.com/moapis/transaction"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	errInternal  = "Internal error"
	errMissingID = "Missing ID"
)

type shopServer struct {
	shop.UnimplementedShopServer

	mdb  *multidb.MultiDB
	log  *logrus.Entry
	conf *ServerConfig
	tv   *transaction.Verificator
	mail *mailer.Mailer
}

func (s *shopServer) SaveArticle(ctx context.Context, req *shop.Article) (*shop.ArticleID, error) {
	rt, err := s.newAuthTx(ctx, "SaveArticle", false, req.GetToken())
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	art, err := rt.upsertArticle(req)
	if err != nil {
		return nil, err
	}
	if err = rt.setArticleCategories(art, req.GetCategories()); err != nil {
		return nil, err
	}
	if err := rt.setArticleBasePrices(art, req.GetBaseprices()); err != nil {
		return nil, err
	}
	if err := rt.updateVariants(art.ID, req.GetVariants()); err != nil {
		return nil, err
	}
	if err := rt.updateImages(art.ID, req.GetImages()); err != nil {
		return nil, err
	}
	if err := rt.updateVideos(art.ID, req.GetVideos()); err != nil {
		return nil, err
	}
	if err = rt.Commit(); err != nil {
		return nil, err
	}
	return &shop.ArticleID{Id: int32(art.ID)}, nil
}

func (s *shopServer) ViewArticle(ctx context.Context, req *shop.ArticleID) (*shop.Article, error) {
	rt, err := s.newTx(ctx, "ViewArticle", true)
	if err != nil {
		return nil, err
	}
	defer rt.Done()
	aid := req.GetId()
	if aid == 0 {
		rt.Log.Warn(errMissingID)
		return nil, status.Error(codes.InvalidArgument, errMissingID)
	}

	return rt.viewArticle(int(aid))
}

func (s *shopServer) ListArticles(ctx context.Context, req *shop.ListConditions) (*shop.ArticleList, error) {
	rt, err := s.newTx(ctx, "ListArticles", true)
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	list, err := rt.listArticles(req)
	if err != nil {
		return nil, err
	}
	return &shop.ArticleList{List: list}, nil
}

func (s *shopServer) DeleteArticle(ctx context.Context, req *shop.ArticleID) (*shop.Deleted, error) {
	rt, err := s.newAuthTx(ctx, "DeleteArticle", false, req.GetToken())
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	aid := req.GetId()
	if aid == 0 {
		rt.Log.Warn(errMissingID)
		return nil, status.Error(codes.InvalidArgument, errMissingID)
	}
	rt.Log = rt.Log.WithField("aid", aid)

	ra, err := rt.deleteArticle(int(aid))
	if err != nil {
		return nil, err
	}

	if err = rt.Commit(); err != nil {
		return nil, err
	}

	return &shop.Deleted{Rows: ra}, nil
}

func (s *shopServer) Checkout(ctx context.Context, req *shop.Order) (*shop.OrderID, error) {
	rt, err := s.newTx(ctx, "Checkout", false)
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	order, err := rt.newOrder(req)
	if err != nil {
		return nil, err
	}

	if err = rt.sendOrderMail(OrderMailTmpl, order, fmt.Sprintf("New order #%d at %s", order.ID, rt.s.conf.Mail.ShopName)); err != nil {
		return nil, err
	}
	encText, encKey, err := rt.encryptOrder(order)
	if err != nil {
		return nil, err
	}
	if err = rt.Commit(); err != nil {
		return nil, err
	}
	return &shop.OrderID{Id: int32(order.ID), EnvKey: encKey, Data: encText}, nil
}

func (s *shopServer) ListOrders(ctx context.Context, req *shop.ListOrderConditions) (*shop.OrderList, error) {
	rt, err := s.newAuthTx(ctx, "ListOrders", true, req.GetToken())
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	return rt.listOrders(req)
}

func (s *shopServer) SaveOrder(ctx context.Context, req *shop.Order) (*shop.OrderID, error) {
	rt, err := s.newAuthTx(ctx, "SaveOrder", false, req.GetToken())
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	order, err := rt.saveOrder(req)
	if err != nil {
		return nil, err
	}

	if err = rt.sendOrderMail(OrderMailTmpl, order, fmt.Sprintf("Update on your order #%d at %s", order.ID, rt.s.conf.Mail.ShopName)); err != nil {
		return nil, err
	}

	if err = rt.Commit(); err != nil {
		return nil, err
	}

	return &shop.OrderID{Id: int32(order.ID)}, nil
}

func (s *shopServer) SaveCategories(ctx context.Context, req *shop.CategoryList) (*shop.CategoryList, error) {
	rt, err := s.newAuthTx(ctx, "SaveCategories", false, req.GetToken())
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	if err = rt.saveCategories(req.GetList()); err != nil {
		return nil, err
	}

	list, err := rt.listCategories(nil)
	if err != nil {
		return nil, err
	}

	if err = rt.Commit(); err != nil {
		return nil, err
	}

	return &shop.CategoryList{List: list}, nil
}

func (s *shopServer) ListCategories(ctx context.Context, req *shop.CategoryListConditions) (*shop.CategoryList, error) {
	rt, err := s.newTx(ctx, "ListCategories", true)
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	list, err := rt.listCategories(req)
	if err != nil {
		return nil, err
	}

	return &shop.CategoryList{List: list}, nil
}

func (s *shopServer) SearchArticles(ctx context.Context, req *shop.TextSearch) (*shop.ArticleList, error) {
	rt, err := s.newTx(ctx, "SearchArticles", true)
	if err != nil {
		return nil, err
	}
	defer rt.Done()
	list, err := rt.searchArticles(req)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *shopServer) Suggest(ctx context.Context, req *shop.TextSearch) (*shop.SuggestionList, error) {
	rt, err := s.newTx(ctx, "Suggest", true)
	if err != nil {
		return nil, err
	}
	defer rt.Done()
	list, err := rt.suggest(req)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *shopServer) SaveBasePrice(ctx context.Context, sbp *shop.BasePrice) (*shop.BasePrice, error) {
	rt, err := s.newAuthTx(ctx, "SaveBasePrice", false, sbp.GetToken())
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	if sbp, err = rt.upsertBasePrice(sbp); err != nil {
		return nil, err
	}

	if err = rt.Commit(); err != nil {
		return nil, err
	}

	return sbp, nil
}

func (s *shopServer) DeleteBasePrice(ctx context.Context, sbp *shop.BasePrice) (*shop.Deleted, error) {
	rt, err := s.newAuthTx(ctx, "DeleteBasePrice", false, sbp.GetToken())
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	del, err := rt.deleteBasePrice(sbp)
	if err != nil {
		return nil, err
	}

	if err = rt.Commit(); err != nil {
		return nil, err
	}

	return del, nil
}

func (s *shopServer) ListBasesPrices(ctx context.Context, bpc *shop.BasePriceListCondtions) (*shop.BasePriceList, error) {
	rt, err := s.newTx(ctx, "ListBasesPrices", true)
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	list, err := rt.listBasePrices(bpc)
	if err != nil {
		return nil, err
	}

	return &shop.BasePriceList{List: list}, err
}

func (s *shopServer) SendMessage(ctx context.Context, sm *shop.Message) (*shop.MessageID, error) {
	rt, err := s.newTx(ctx, "SendMessage", false)
	if err != nil {
		return nil, err
	}
	defer rt.Done()

	msg, err := rt.newMessage(sm)
	if err != nil {
		return nil, err
	}

	if err = rt.sendMail(MessageMailTmpl, msg); err != nil {
		return nil, err
	}

	rt.Commit()

	return &shop.MessageID{Id: int32(msg.ID)}, nil
}
