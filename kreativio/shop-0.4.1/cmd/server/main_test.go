package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
	auth "github.com/moapis/authenticator"
	"github.com/moapis/multidb"
	"github.com/moapis/shop/models"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
)

const (
	migrationsDir = "../../migrations"
)

func migrations() {
	migrate.SetTable("shop_migrations")
	m, err := mdb.Master(testCtx)
	if err != nil {
		log.WithError(err).Fatal("migrations()")
	}
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}
	n, err := migrate.Exec(m.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		log.WithError(err).Fatal("Migrations")
	}
	log.WithField("n", n).Info("Migrations")
}

func migrateDown() {
	m, err := mdb.Master(testCtx)
	if err != nil {
		log.WithError(err).Fatal("migrateDown")
	}
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}
	n, err := migrate.Exec(m.DB, "postgres", migrations, migrate.Down)
	if err != nil {
		log.WithError(err).Fatal("migrateDown")
	}
	log.WithField("n", n).Info("migrateDown")
}

var (
	testCategories = []*models.Category{
		{
			ID:        20,
			CreatedAt: time.Unix(4000, 0),
			UpdatedAt: time.Unix(5000, 0),
			Label:     "Empty category",
			Position:  1,
		},
		{
			ID:        21,
			CreatedAt: time.Unix(6000, 0),
			UpdatedAt: time.Unix(7000, 0),
			Label:     "Full category",
			Position:  2,
		},
		{
			ID:        22,
			CreatedAt: time.Unix(8000, 0),
			UpdatedAt: time.Unix(9000, 0),
			Label:     "Only unpublished category",
			Position:  3,
		},
	}
	testBasePrices = []*models.BasePrice{
		{
			ID:        31,
			CreatedAt: time.Unix(4000, 0),
			UpdatedAt: time.Unix(5000, 0),
			Label:     "Cheap material",
			Price:     types.NewDecimal(decimal.New(4455, 2)),
		},
		{
			ID:        32,
			CreatedAt: time.Unix(6000, 0),
			UpdatedAt: time.Unix(7000, 0),
			Label:     "Premium material",
			Price:     types.NewDecimal(decimal.New(9999, 2)),
		},
	}
	testArticles = []*models.Article{
		{
			ID:          11,
			CreatedAt:   time.Unix(8000, 0),
			UpdatedAt:   time.Unix(9000, 0),
			Published:   false,
			Title:       "ID 11",
			Description: "This is the first article",
			Price:       types.NewDecimal(decimal.New(3000099, 2)),
			Promoted:    false,
		},
		{
			ID:          12,
			CreatedAt:   time.Unix(1000, 0),
			UpdatedAt:   time.Unix(2000, 0),
			Published:   true,
			Title:       "ID 12",
			Description: "This is the second article",
			Price:       types.NewDecimal(decimal.New(1212, 2)),
			Promoted:    false,
		},
		{
			ID:          13,
			CreatedAt:   time.Unix(3000, 0),
			UpdatedAt:   time.Unix(4000, 0),
			Published:   false,
			Title:       "ID 13",
			Description: "This is the third article",
			Price:       types.NewDecimal(decimal.New(2299, 2)),
			Promoted:    true,
		},
	}
	testVariants = []*models.Variant{
		{
			ID:         41,
			ArticleID:  13,
			Labels:     []string{"hello", "world"},
			Multiplier: types.NewDecimal(decimal.New(333, 2)),
		},
		{
			ID:         42,
			ArticleID:  13,
			Labels:     []string{"foo", "bar"},
			Multiplier: types.NewDecimal(decimal.New(999, 2)),
		},
	}
	testImages = []*models.Image{
		{
			ID:        111,
			ArticleID: 11,
			Position:  1,
			Label:     "First image on first article",
			URL:       "https://bucket.s3.com/firstonfirst.jpg",
		},
		{
			ID:        112,
			ArticleID: 11,
			Position:  2,
			Label:     "Second image on first article",
			URL:       "https://bucket.s3.com/secondonfirst.jpg",
		},
		{
			ID:        113,
			ArticleID: 11,
			Position:  3,
			Label:     "Third image on first article",
			URL:       "https://bucket.s3.com/thirdonfirst.jpg",
		},
		{
			ID:        121,
			ArticleID: 12,
			Position:  1,
			Label:     "First image on second article",
			URL:       "https://bucket.s3.com/firstonsecond.jpg",
		},
		{
			ID:        122,
			ArticleID: 12,
			Position:  2,
			Label:     "Second image on second article",
			URL:       "https://bucket.s3.com/secondonsecond.jpg",
		},
		{
			ID:        123,
			ArticleID: 12,
			Position:  3,
			Label:     "Third image on second article",
			URL:       "https://bucket.s3.com/thirdonsecond.jpg",
		},
		{
			ID:        131,
			ArticleID: 13,
			Position:  1,
			Label:     "First image on third article",
			URL:       "https://bucket.s3.com/firstonthird.jpg",
		},
		{
			ID:        132,
			ArticleID: 13,
			Position:  2,
			Label:     "Second image on third article",
			URL:       "https://bucket.s3.com/secondonthird.jpg",
		},
		{
			ID:        133,
			ArticleID: 13,
			Position:  3,
			Label:     "Third image on third article",
			URL:       "https://bucket.s3.com/thirdonthird.jpg",
		},
	}
	testVideos = []*models.Video{
		{
			ID:        111,
			ArticleID: 11,
			Position:  1,
			Label:     "First video on first article",
			URL:       "https://bucket.s3.com/firstonfirst.mp4",
		},
		{
			ID:        112,
			ArticleID: 11,
			Position:  2,
			Label:     "Second video on first article",
			URL:       "https://bucket.s3.com/secondonfirst.mp4",
		},
		{
			ID:        113,
			ArticleID: 11,
			Position:  3,
			Label:     "Third video on first article",
			URL:       "https://bucket.s3.com/thirdonfirst.mp4",
		},
		{
			ID:        121,
			ArticleID: 12,
			Position:  1,
			Label:     "First video on second article",
			URL:       "https://bucket.s3.com/firstonsecond.mp4",
		},
		{
			ID:        122,
			ArticleID: 12,
			Position:  2,
			Label:     "Second video on second article",
			URL:       "https://bucket.s3.com/secondonsecond.mp4",
		},
		{
			ID:        123,
			ArticleID: 12,
			Position:  3,
			Label:     "Third video on second article",
			URL:       "https://bucket.s3.com/thirdonsecond.mp4",
		},
		{
			ID:        131,
			ArticleID: 13,
			Position:  1,
			Label:     "First video on third article",
			URL:       "https://bucket.s3.com/firstonthird.mp4",
		},
		{
			ID:        132,
			ArticleID: 13,
			Position:  2,
			Label:     "Second video on third article",
			URL:       "https://bucket.s3.com/secondonthird.mp4",
		},
		{
			ID:        133,
			ArticleID: 13,
			Position:  3,
			Label:     "Third video on third article",
			URL:       "https://bucket.s3.com/thirdonthird.mp4",
		},
	}
	testOrders = []*models.Order{
		{
			ID:            100,
			CreatedAt:     time.Unix(12, 0),
			UpdatedAt:     time.Unix(34, 0),
			FullName:      "What Is My Name",
			Email:         "me@example.com",
			Phone:         "0123456789",
			FullAddress:   "No. 7, Long Street, Somewhere",
			Message:       "My awesome order",
			PaymentMethod: models.PaymentCASH_ON_DELIVERY,
			Status:        models.StatusSENT,
		},
		{
			ID:            101,
			CreatedAt:     time.Unix(56, 0),
			UpdatedAt:     time.Unix(78, 0),
			FullName:      "What Is My Name",
			Email:         "me@example.com",
			Phone:         "0123456789",
			FullAddress:   "No. 7, Long Street, Somewhere",
			Message:       "Undefined status",
			PaymentMethod: models.PaymentCASH_ON_DELIVERY,
			Status:        models.StatusUNDEFINED,
		},
	}
	testOrderArticles = []*models.OrderArticle{
		{
			OrderID:   100,
			ArticleID: 11,
			Amount:    1,
			ID:        601,
			Title:     "ID 11",
			Price:     types.NewDecimal(decimal.New(3000099, 2)),
		},
		{
			OrderID:   100,
			ArticleID: 12,
			Amount:    3,
			ID:        602,
			Title:     "ID 12",
			Price:     types.NewDecimal(decimal.New(1212, 2)),
		},
		{
			OrderID:   100,
			ArticleID: 13,
			Amount:    5,
			ID:        603,
			Title:     "ID 13",
			Price:     types.NewDecimal(decimal.New(1483515, 4)),
			Details: null.NewJSON(
				[]byte(`{
					"base_price":{"label":"Cheap material", "price":"44.55"},
					"variant":{"labels": ["hello", "world"], "multiplier":"3.33"}
				}`), true,
			),
		},
		{
			OrderID:   101,
			ArticleID: 11,
			Amount:    22,
			ID:        604,
			Title:     "ID 11",
			Price:     types.NewDecimal(decimal.New(2299, 2)),
		},
	}
)

func testData() error {
	tx, err := mdb.MasterTx(testCtx, nil)
	if err != nil {
		log.WithError(err).Error("Open TX")
		return err
	}
	defer tx.Rollback()

	for _, cat := range testCategories {
		log := log.WithField("category", cat)
		if err = cat.Insert(testCtx, tx, boil.Infer()); err != nil {
			log.WithError(err).Error("categories")
			return err
		}
	}

	for _, art := range testArticles {
		log := log.WithField("article", art)
		if err = art.Insert(testCtx, tx, boil.Infer()); err != nil {
			log.WithError(err).Error("articles")
			return err
		}
		switch art.Published {
		case true:
			err = art.SetCategories(testCtx, tx, false, testCategories[1])
		case false:
			err = art.SetCategories(testCtx, tx, false, testCategories[1], testCategories[2])
		}
		if err != nil {
			log.WithError(err).Error("art.SetCategories")
			return err
		}

		log.Debug("testData articles")
	}

	if err := testArticles[2].SetBasePrices(testCtx, tx, true, testBasePrices...); err != nil {
		log.WithError(err).Error("basePrices")
		return err
	}
	log.WithField("basePrices", testBasePrices).Debug("basePrices")

	for _, vrt := range testVariants {
		log := log.WithField("variant", vrt)
		if err = vrt.Insert(testCtx, tx, boil.Infer()); err != nil {
			log.WithError(err).Error("variants")
			return err
		}
		log.Debug("testData variants")
	}

	for _, img := range testImages {
		log := log.WithField("image", img)
		if err = img.Insert(testCtx, tx, boil.Infer()); err != nil {
			log.WithError(err).Error("images")
			return err
		}
		log.Debug("testData images")
	}
	for _, vid := range testVideos {
		log := log.WithField("video", vid)
		if err = vid.Insert(testCtx, tx, boil.Infer()); err != nil {
			log.WithError(err).Error("videos")
			return err
		}
		log.Debug("testData videos")
	}
	for _, ord := range testOrders {
		log := log.WithField("order", ord)
		if err = ord.Insert(testCtx, tx, boil.Infer()); err != nil {
			log.WithError(err).Error("order")
			return err
		}
	}
	for _, oa := range testOrderArticles {
		log := log.WithField("order_article", oa)
		if err = oa.Insert(testCtx, tx, boil.Infer()); err != nil {
			log.WithError(err).Error("order_article")
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		log.WithError(err).Error("commit")
		return err
	}
	log.Info("testData inserted")
	return nil
}

var (
	testConfig *ServerConfig
	testCtx    context.Context
	mdb        *multidb.MultiDB
	tss        *shopServer
)

func TestMain(m *testing.M) {
	var err error
	testConfig, err = configure(Default)
	if err != nil {
		log.WithError(err).Fatal("configure()")
	}

	var cancel context.CancelFunc
	testCtx, cancel = context.WithTimeout(context.Background(), 30*time.Second)

	mdb, err = testConfig.MultiDB.Open()
	if err != nil {
		log.WithError(err).Fatal("mdb.Open()")
	}

	migrations()
	if err = testData(); err != nil {
		migrateDown()
		log.WithError(err).Fatal("testData")
	}

	if tss, err = testConfig.newShopServer(); err != nil {
		migrateDown()
		log.WithError(err).Fatal("newShopServer")
	}

	ar, err := tss.tv.Client.AuthenticatePwUser(
		testCtx,
		&auth.UserPassword{
			Email:    "admin@localhost",
			Password: "admin",
		},
	)
	if err != nil {
		migrateDown()
		log.WithError(err).Fatal("AuthenticatePwUser")
	}
	testToken = ar.GetJwt()

	code := m.Run()

	migrateDown()
	cancel()
	os.Exit(code)
}
