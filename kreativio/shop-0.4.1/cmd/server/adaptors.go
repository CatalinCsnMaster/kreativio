package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/moapis/shop"
	"github.com/moapis/shop/models"
	fj "github.com/valyala/fastjson"
	"github.com/volatiletech/sqlboiler/v4/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	errMissing   = "Missing required fields: %s"
	errNoArts    = "No articles in order"
	errNegAmount = "Negative amount on order article %d: %d"
	errEnum      = "ENUM mismatch: %v"
	errDecimal   = "Can't convert %s string %s to decimal" // Field name and value
)

func checkRequired(vals map[string]interface{}) error {
	var empty []string
	for k, v := range vals {
		if v == nil || v == "" || v == 0 || v == 0.0 {
			empty = append(empty, k)
		} else if ss, ok := v.([]string); ok && len(ss) == 0 {
			empty = append(empty, k)
		}
	}

	sort.Strings(empty)

	if len(empty) > 0 {
		return status.Errorf(
			codes.InvalidArgument,
			errMissing,
			strings.Join(empty, ", "),
		)
	}
	return nil
}

const (
	artDecimal = "Price"
)

func articleMsgToModel(sa *shop.Article) (*models.Article, error) {
	vals := map[string]interface{}{
		"Title":       sa.GetTitle(),
		"Description": sa.GetDescription(),
		artDecimal:    sa.GetPrice(),
	}
	if err := checkRequired(vals); err != nil {
		return nil, err
	}

	price, ok := new(decimal.Big).SetString(vals[artDecimal].(string))
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, errDecimal, artDecimal, vals[artDecimal])
	}

	return &models.Article{
		ID:          int(sa.GetId()),
		Published:   sa.GetPublished(),
		Title:       vals["Title"].(string),
		Description: vals["Description"].(string),
		Price:       types.NewDecimal(price),
		Promoted:    sa.GetPromoted(),
	}, nil
}

const (
	varDecimal = "Multiplier"
)

func variantsMsgToModel(aid int, sv []*shop.Variant) ([]*models.Variant, error) {
	vars := make([]*models.Variant, len(sv))
	for i, v := range sv {
		vals := map[string]interface{}{
			"ArticleID": aid,
			"Labels":    v.GetLabels(),
			varDecimal:  v.GetMultiplier(),
		}
		if err := checkRequired(vals); err != nil {
			return nil, err
		}

		mpl, ok := new(decimal.Big).SetString(vals[varDecimal].(string))
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, errDecimal, varDecimal, vals[varDecimal])
		}

		vars[i] = &models.Variant{
			ID:         v.GetId(),
			ArticleID:  vals["ArticleID"].(int),
			Labels:     vals["Labels"].([]string),
			Multiplier: types.NewDecimal(mpl),
		}
	}

	return vars, nil
}

func imagesMsgToModel(aid int, sm []*shop.Media) ([]*models.Image, error) {
	imgs := make([]*models.Image, len(sm))
	for i, m := range sm {
		req := map[string]interface{}{
			"ArticleID": aid,
			"Label":     m.GetLabel(),
			"URL":       m.GetUrl(),
		}
		if err := checkRequired(req); err != nil {
			return nil, err
		}
		imgs[i] = &models.Image{
			ID:        int(m.GetId()),
			ArticleID: req["ArticleID"].(int),
			Position:  i + 1,
			Label:     req["Label"].(string),
			URL:       req["URL"].(string),
		}
	}
	return imgs, nil
}

func videosMsgToModel(aid int, sm []*shop.Media) ([]*models.Video, error) {
	vids := make([]*models.Video, len(sm))
	for i, m := range sm {
		req := map[string]interface{}{
			"ArticleID": aid,
			"Label":     m.GetLabel(),
			"URL":       m.GetUrl(),
		}
		if err := checkRequired(req); err != nil {
			return nil, err
		}
		vids[i] = &models.Video{
			ID:        int(m.GetId()),
			ArticleID: req["ArticleID"].(int),
			Position:  i + 1,
			Label:     req["Label"].(string),
			URL:       req["URL"].(string),
		}
	}
	return vids, nil
}

func timeModelToMsg(created, updated time.Time) (c *timestamp.Timestamp, u *timestamp.Timestamp, err error) {
	if !created.IsZero() {
		c, err = ptypes.TimestampProto(created)
	}
	if err != nil {
		return nil, nil, status.Error(codes.OutOfRange, err.Error())
	}

	if !updated.IsZero() {
		u, err = ptypes.TimestampProto(updated)
	}
	if err != nil {
		return nil, nil, status.Error(codes.OutOfRange, err.Error())
	}
	return c, u, nil
}

func articleModelToMsg(art *models.Article) (*shop.Article, error) {
	created, updated, err := timeModelToMsg(art.CreatedAt, art.UpdatedAt)
	if err != nil {
		return nil, err
	}

	sa := &shop.Article{
		Id:          int32(art.ID),
		Created:     created,
		Updated:     updated,
		Published:   art.Published,
		Title:       art.Title,
		Description: art.Description,
		Price:       art.Price.String(),
		Promoted:    art.Promoted,
	}

	if art.R != nil {
		if art.R.Images != nil {
			sa.Images = make([]*shop.Media, len(art.R.Images))
			for i, img := range art.R.Images {
				sa.Images[i] = &shop.Media{
					Id:    int32(img.ID),
					Label: img.Label,
					Url:   img.URL,
				}
			}
		}

		if art.R.Videos != nil {
			sa.Videos = make([]*shop.Media, len(art.R.Videos))
			for i, vid := range art.R.Videos {
				sa.Videos[i] = &shop.Media{
					Id:    int32(vid.ID),
					Label: vid.Label,
					Url:   vid.URL,
				}
			}
		}

		if art.R.Categories != nil {
			sa.Categories = make([]*shop.Category, len(art.R.Categories))
			for i, cat := range art.R.Categories {
				sa.Categories[i] = &shop.Category{
					Id:    int32(cat.ID),
					Label: cat.Label,
				}
			}
		}

		if art.R.BasePrices != nil {
			sa.Baseprices = make([]*shop.BasePrice, len(art.R.BasePrices))
			for i, bp := range art.R.BasePrices {
				sa.Baseprices[i] = &shop.BasePrice{
					Id:    int32(bp.ID),
					Label: bp.Label,
					Price: bp.Price.String(),
				}
			}
		}

		if art.R.Variants != nil {
			sa.Variants = make([]*shop.Variant, len(art.R.Variants))
			for i, vrt := range art.R.Variants {
				sa.Variants[i] = &shop.Variant{
					Id:         vrt.ID,
					Labels:     vrt.Labels,
					Multiplier: vrt.Multiplier.String(),
				}
			}
		}
	}
	return sa, nil
}

func orderMsgToModel(so *shop.Order) (*models.Order, error) {
	vals := map[string]interface{}{
		"FullName":    so.GetFullName(),
		"Email":       so.GetEmail(),
		"Phone":       so.GetPhone(),
		"FullAddress": so.GetFullAddress(),
	}
	if err := checkRequired(vals); err != nil {
		return nil, err
	}

	return &models.Order{
		FullName:      vals["FullName"].(string),
		Email:         vals["Email"].(string),
		Phone:         vals["Phone"].(string),
		FullAddress:   vals["FullAddress"].(string),
		Message:       so.GetMessage(),
		PaymentMethod: so.GetPaymentMethod().String(),
	}, nil
}

func checkOrderArticles(sa []*shop.Order_ArticleAmount) error {
	var total int

	for _, a := range sa {
		aid := int(a.GetArticleId())
		amt := int(a.GetAmount())
		if a.Amount < 0 {
			return status.Errorf(codes.InvalidArgument, errNegAmount, aid, amt)
		}
		total += amt
	}
	if total == 0 {
		return status.Error(codes.InvalidArgument, errNoArts)
	}

	return nil
}

func orderArticlesModelsToMsg(arts []*models.OrderArticle) ([]*shop.Order_ArticleAmount, string, error) {
	sum := new(decimal.Big)
	soaa := make([]*shop.Order_ArticleAmount, len(arts))
	for i, a := range arts {
		total := new(decimal.Big).Mul(
			a.Price.Big,
			decimal.New(int64(a.Amount), 0),
		)

		soaa[i] = &shop.Order_ArticleAmount{
			ArticleId: int32(a.ArticleID),
			Amount:    int32(a.Amount),
			Title:     a.Title,
			Price:     a.Price.String(),
			Total:     total.String(),
		}
		sum.Add(sum, total)

		if a.Details.Valid {
			soaa[i].Details = new(shop.Details)
			if err := json.Unmarshal(a.Details.JSON, soaa[i].Details); err != nil {
				return nil, "", err
			}
		}
	}
	return soaa, sum.String(), nil
}

func orderModelToMsg(order *models.Order) (*shop.Order, error) {
	created, updated, err := timeModelToMsg(order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	pm, ok := shop.Order_PaymentMethod_value[order.PaymentMethod]
	if !ok {
		return nil, status.Errorf(codes.Unimplemented, errEnum, order.PaymentMethod)
	}
	os, ok := shop.Order_Status_value[order.Status]
	if !ok {
		return nil, status.Errorf(codes.Unimplemented, errEnum, order.Status)
	}

	return &shop.Order{
		Id:            int32(order.ID),
		Created:       created,
		Updated:       updated,
		FullName:      order.FullName,
		Email:         order.Email,
		Phone:         order.Phone,
		FullAddress:   order.FullAddress,
		Message:       order.Message,
		PaymentMethod: shop.Order_PaymentMethod(pm),
		Status:        shop.Order_Status(os),
	}, nil
}

func orderUpdateMsgToModel(so *shop.Order) (*models.Order, error) {
	order := models.Order{
		ID:            int(so.GetId()),
		FullName:      so.GetFullName(),
		Email:         so.GetEmail(),
		Phone:         so.GetPhone(),
		FullAddress:   so.GetFullAddress(),
		Message:       so.GetMessage(),
		PaymentMethod: so.GetPaymentMethod().String(),
		Status:        so.GetStatus().String(),
	}
	if order.ID <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, errMissing, "ID")
	}
	return &order, nil
}

func categoriesMsgToModel(list []*shop.Category) ([]*models.Category, error) {
	cats := make([]*models.Category, len(list))
	for i, c := range list {
		cats[i] = &models.Category{
			ID:       int(c.GetId()),
			Label:    c.GetLabel(),
			Position: i + 1,
		}
		if cats[i].Label == "" {
			return nil, status.Errorf(codes.InvalidArgument, errMissing, "Label")
		}
	}
	return cats, nil
}

func categoriesModeltoMsg(cats []*models.Category) ([]*shop.Category, error) {
	scs := make([]*shop.Category, len(cats))

	for i, c := range cats {
		created, updated, err := timeModelToMsg(c.CreatedAt, c.UpdatedAt)
		if err != nil {
			return nil, err
		}

		scs[i] = &shop.Category{
			Id:      int32(c.ID),
			Created: created,
			Updated: updated,
			Label:   c.Label,
		}
	}

	return scs, nil
}

// TODO: modify query to return EPOCH seconds directly
func timeBytesToMsg(ca, ua []byte) (created, updated *timestamp.Timestamp, err error) {
	var (
		ct, ut time.Time
	)

	if len(ca) != 0 {
		if ct, err = time.Parse(time.RFC3339, string(ca)); err != nil {
			return nil, nil, status.Error(codes.OutOfRange, err.Error())
		}
	}
	if len(ua) != 0 {
		if ut, err = time.Parse(time.RFC3339, string(ua)); err != nil {
			return nil, nil, status.Error(codes.OutOfRange, err.Error())
		}
	}

	return timeModelToMsg(ct, ut)
}

func mediaValuesToMsg(mds []*fj.Value) []*shop.Media {
	if len(mds) == 0 {
		return nil
	}
	sm := make([]*shop.Media, len(mds))

	for i, m := range mds {
		sm[i] = &shop.Media{
			Id:    int32(m.GetInt("id")),
			Label: string(m.GetStringBytes("label")),
			Url:   string(m.GetStringBytes("url")),
		}
	}

	return sm
}

func categoryValuesToMsg(cvs []*fj.Value) []*shop.Category {
	if len(cvs) == 0 {
		return nil
	}
	sc := make([]*shop.Category, len(cvs))

	for i, c := range cvs {
		sc[i] = &shop.Category{
			Id:    int32(c.GetInt("id")),
			Label: string(c.GetStringBytes("label")),
		}
	}
	return sc
}

func basePriceValuesToMsg(bpv []*fj.Value) []*shop.BasePrice {
	if len(bpv) == 0 {
		return nil
	}
	sbp := make([]*shop.BasePrice, len(bpv))

	for i, b := range bpv {
		sbp[i] = &shop.BasePrice{
			Id:    int32(b.GetInt("id")),
			Label: string(b.GetStringBytes("label")),
			Price: string(b.GetStringBytes("price")),
		}
	}
	return sbp
}

func variantValuesToMsg(vts []*fj.Value) []*shop.Variant {
	if len(vts) == 0 {
		return nil
	}
	sv := make([]*shop.Variant, len(vts))

	for i, v := range vts {
		sv[i] = &shop.Variant{
			Id:         v.GetInt64("id"),
			Multiplier: string(v.GetStringBytes("multiplier")),
		}
		for _, l := range v.GetArray("labels") {
			s, _ := strconv.Unquote(l.String())
			sv[i].Labels = append(sv[i].Labels, s)
		}
	}
	return sv
}

func articleValueToMsg(art *fj.Value) (sa *shop.Article, err error) {
	sa = &shop.Article{
		Id:          int32(art.GetInt("id")),
		Published:   art.GetBool("published"),
		Title:       string(art.GetStringBytes("title")),
		Description: string(art.GetStringBytes("description")),
		Price:       string(art.GetStringBytes("price")),
		Promoted:    art.GetBool("promoted"),
		Images:      mediaValuesToMsg(art.GetArray("images")),
		Videos:      mediaValuesToMsg(art.GetArray("videos")),
		Categories:  categoryValuesToMsg(art.GetArray("categories")),
		Baseprices:  basePriceValuesToMsg(art.GetArray("baseprices")),
		Variants:    variantValuesToMsg(art.GetArray("variants")),
	}

	if sa.Created, sa.Updated, err = timeBytesToMsg(art.GetStringBytes("created_at"), art.GetStringBytes("updated_at")); err != nil {
		return nil, err
	}

	return sa, nil
}

type jsonScanner struct {
	*fj.Value
	p *fj.Parser
}

var (
	errJSONEmpty  = errors.New("JSON result empty")
	errJSONAssert = errors.New("JSON assertion error")
)

func (s *jsonScanner) Scan(v interface{}) (err error) {
	if v == nil {
		return errJSONEmpty
	}

	js, ok := v.([]byte)
	if !ok {
		return errJSONAssert
	}

	if s.Value, err = s.p.ParseBytes(js); err != nil {
		return fmt.Errorf("json Scan: %w", err)
	}

	return nil
}

func (s *jsonScanner) articlesJSONtoMsg() ([]*shop.Article, error) {
	va, err := s.Array()
	if err != nil {
		return nil, status.Error(codes.Internal, errFatal)
	}

	sas := make([]*shop.Article, len(va))
	for i, a := range va {
		if sas[i], err = articleValueToMsg(a); err != nil {
			return nil, err
		}
	}
	return sas, nil
}

func basePriceMsgToModel(sbp *shop.BasePrice) (*models.BasePrice, error) {
	vals := map[string]interface{}{
		"Label": sbp.GetLabel(),
		"Price": sbp.GetPrice(),
	}
	if err := checkRequired(vals); err != nil {
		return nil, err
	}

	price, ok := new(decimal.Big).SetString(vals["Price"].(string))
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, errDecimal, artDecimal, vals["Price"])
	}

	return &models.BasePrice{
		ID:    int(sbp.GetId()),
		Label: vals["Label"].(string),
		Price: types.NewDecimal(price),
	}, nil
}

func basePriceModeltoMsg(bp *models.BasePrice) (*shop.BasePrice, error) {
	created, updated, err := timeModelToMsg(bp.CreatedAt, bp.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &shop.BasePrice{
		Id:      int32(bp.ID),
		Created: created,
		Updated: updated,
		Label:   bp.Label,
		Price:   bp.Price.String(),
	}, nil
}

func basePricesModeltoMsg(bps []*models.BasePrice) ([]*shop.BasePrice, error) {
	sbp := make([]*shop.BasePrice, len(bps))

	for i, bp := range bps {
		var err error
		if sbp[i], err = basePriceModeltoMsg(bp); err != nil {
			return nil, err
		}
	}

	return sbp, nil
}

func messageMsgToModel(sm *shop.Message) (*models.Message, error) {
	vals := map[string]interface{}{
		"Name":    sm.GetName(),
		"Email":   sm.GetEmail(),
		"Message": sm.GetMessage(),
	}
	if err := checkRequired(vals); err != nil {
		return nil, err
	}

	return &models.Message{
		Name:    vals["Name"].(string),
		Email:   vals["Email"].(string),
		Message: vals["Message"].(string),
		Phone:   sm.GetPhone(),
		Subject: sm.GetSubject(),
	}, nil
}
