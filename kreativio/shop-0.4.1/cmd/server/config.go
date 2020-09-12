// Copyright (c) 2019, Mohlmann Solutions SRL. All rights reserved.
// Use of this source code is governed by a License that can be found in the LICENSE file.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/moapis/mailer"
	"github.com/moapis/multidb"
	pg "github.com/moapis/multidb/drivers/postgresql"
	"github.com/moapis/shop"
	"github.com/moapis/shop/builder"
	"github.com/moapis/shop/mobilpay"
	"github.com/moapis/transaction"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	log *logrus.Logger
)

// LogLevel used for logrus
type LogLevel string

const (
	// PanicLevel sets logrus level to panic
	PanicLevel LogLevel = "panic"
	// FatalLevel sets logrus level to fatal
	FatalLevel LogLevel = "fatal"
	// ErrorLevel sets logrus level to error
	ErrorLevel LogLevel = "error"
	// WarnLevel sets logrus level to warn
	WarnLevel LogLevel = "warn"
	// InfoLevel sets logrus level to info
	InfoLevel LogLevel = "info"
	// DebugLevel sets logrus level to debug
	DebugLevel LogLevel = "debug"
	// TraceLevel sets logrus level to trace
	TraceLevel LogLevel = "trace"
)

func init() {
	log = logrus.New()
	log.SetLevel(logrus.InfoLevel)
}

// TLSConfig for the gRPC server's CertFile and KeyFile
type TLSConfig struct {
	CertFile string `json:"certfile,omitempty"`
	KeyFile  string `json:"keyfile,omitempty"`
}

// AuthServerConfig for the gRPC client connection
type AuthServerConfig struct {
	Host string
	Port uint16
}

func (acs AuthServerConfig) String() string {
	return fmt.Sprintf("%s:%d", acs.Host, acs.Port)
}

// MailConfig for outgoing mail server
type MailConfig struct {
	Host         string
	Port         uint16
	Identity     string
	Username     string
	Password     string
	From         string
	To           []string
	TemplateGlob string
	ShopName     string
	Currency     string
}
type httpServer struct {
	Address          string
	MobilpayEndpoint string
}
type mobilpayCfg struct {
	CertificateFile string
	PrivateKeyFile  string
	Signature       string
	ConfirmURL      string
	ReturnURL       string
}

// ServerConfig is a collection of config
type ServerConfig struct {
	Addres      string              `json:"address"`     // gRPC listen Address
	Port        uint16              `json:"port"`        // gRPC listen Port
	LogLevel    LogLevel            `json:"loglevel"`    // LogLevel used for logrus
	TLS         *TLSConfig          `json:"tls"`         // TLS will be disabled when nil
	AuthServer  AuthServerConfig    `json:"authserver"`  // Config for the gRPC client connection
	Audiences   []string            `json:"audiences"`   // Accepted audiences from JWT
	Groups      map[string][]string `json:"groups"`      // Map of method names and allowed user groups
	MultiDB     multidb.Config      `json:"multidb"`     // Imported from multidb
	PG          *pg.Config          `json:"pg"`          // PG is later embedded in multidb
	SQLRoutines int                 `json:"sqlroutines"` // Amount of Go-routines for non-master queries
	Mail        MailConfig          `json:"smtp"`
	HTTPServer  httpServer          `json:"http"`
	Mobilpay    mobilpayCfg         `json:"mobilpay"`
	ListLimit   int32               `json:"list_limit"` // Default limit for List Queries, when ommited in the ListConditions
}

func (c *ServerConfig) writeOut(filename string) error {
	out, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, out, 0644)
}

// Default config
var Default = ServerConfig{
	Addres:   "127.0.0.1",
	Port:     8766,
	LogLevel: WarnLevel,
	TLS:      nil,
	Groups: map[string][]string{
		"SaveArticle":   {"primary"},
		"DeleteArticle": {"primary"},
		"ListOrders":    {"primary"},
		"SaveOrder":     {"primary"},
	},
	AuthServer: AuthServerConfig{"127.0.0.1", 8765},
	MultiDB: multidb.Config{
		StatsLen:      100,
		MaxFails:      20,
		ReconnectWait: 1 * time.Second,
	},
	PG: &pg.Config{
		Nodes: []pg.Node{
			{
				Host: "localhost",
				Port: 5432,
			},
		},
		Params: pg.Params{
			DBname:          "shop_test",
			User:            "postgres",
			Password:        "",
			SSLmode:         "disable",
			Connect_timeout: 30,
		},
	},
	SQLRoutines: 3,
	Mail: MailConfig{
		Host:         "test.mailu.io",
		Port:         587,
		Identity:     "",
		Username:     "admin@test.mailu.io",
		Password:     "letmein",
		From:         "admin@test.mailu.io",
		TemplateGlob: "templates/*.mail.html",
		To:           []string{"admin@test.mailu.io"},
		ShopName:     "moapis/shop unit tests",
		Currency:     "EUR",
	},
	HTTPServer: httpServer{
		Address:          "0.0.0.0:8080",
		MobilpayEndpoint: "http://sandboxsecure.mobilpay.ro",
	},
	Mobilpay: mobilpayCfg{
		CertificateFile: "config/sandbox.LK1F-GMV1-YWRD-7J6T-QD55.public.cer",
		PrivateKeyFile:  "config/sandbox.LK1F-GMV1-YWRD-7J6T-QD55private.key",
		ReturnURL:       "https://kreativio.ro/sent",
		ConfirmURL:      "https://pay.kreativio.ro/pay/mobilpayConfirm",
		Signature:       "LK1F-GMV1-YWRD-7J6T-QD55",
	},
	ListLimit: builder.DefaultLimit,
}

var configFiles = flag.String("config", "", "Comma separated list of JSON config files")

func configure(c ServerConfig) (*ServerConfig, error) {
	flag.Parse()

	files := strings.Split(*configFiles, ",")
	s := &c
	for _, f := range files {
		if f == "" {
			continue
		}
		log := log.WithField("file", f)
		js, err := ioutil.ReadFile(f)
		if err != nil {
			log.WithError(err).Error("Read config file")
			return nil, err
		}
		if err = json.Unmarshal(js, s); err != nil {
			log.WithError(err).Error("Unmarshal config file")
			return nil, err
		}
		log.Info("Applied config")
	}

	if s.PG != nil {
		s.MultiDB.DBConf, s.PG = s.PG, nil
	}

	if s.LogLevel == DebugLevel {
		boil.DebugMode = true
		mailer.Debug = true
		builder.Debug = true
	} else {
		boil.DebugMode = false
		mailer.Debug = false
		builder.Debug = false
	}

	lvl, err := logrus.ParseLevel(string(s.LogLevel))
	if err != nil {
		return nil, err
	}
	log.WithField("level", lvl).Info("Setting log level")
	log.SetLevel(lvl)

	log.WithField("config", *s).Debug("Config loaded")

	builder.DefaultLimit = s.ListLimit

	return s, nil
}

func (c ServerConfig) grpcOpts() ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption
	if c.TLS != nil {
		log := log.WithFields(logrus.Fields{"certFile": c.TLS.CertFile, "keyFile": c.TLS.KeyFile})
		cert, err := tls.LoadX509KeyPair(c.TLS.CertFile, c.TLS.KeyFile)
		if err != nil {
			log.WithError(err).Error("Failed to set TLS opts")
			return nil, err
		}
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
	}
	return opts, nil
}

// Outgoing mail template names
const (
	OrderMailTmpl   = "checkout"
	MessageMailTmpl = "message"
)

func (c ServerConfig) newShopServer() (*shopServer, error) {
	s := &shopServer{
		log:  log.WithField("server", "Shop"),
		conf: &c,
	}

	tmpl, err := template.ParseGlob(c.Mail.TemplateGlob)
	if err != nil {
		return nil, err
	}
	s.mail = mailer.New(
		tmpl,
		fmt.Sprintf("%s:%d", c.Mail.Host, c.Mail.Port),
		c.Mail.From,
		smtp.PlainAuth(c.Mail.Identity, c.Mail.Username, c.Mail.Password, c.Mail.Host),
	)

	if s.tv, err = transaction.NewVerificator(context.TODO(), s.log, c.AuthServer.String(), c.Audiences...); err != nil {
		return nil, err
	}

	if s.mdb, err = c.MultiDB.Open(); err != nil {
		return nil, err
	}
	return s, nil
}

func (c ServerConfig) listenAndServe(s *shopServer, opts ...grpc.ServerOption) (*grpc.Server, <-chan error) {
	gs := grpc.NewServer(opts...)
	ec := make(chan error)
	shop.RegisterShopServer(gs, s)

	log := s.log.WithFields(logrus.Fields{"address": c.Addres, "port": c.Port})
	log.WithField("grpc", gs.GetServiceInfo()).Debug("Registered services")
	log.Info("Starting server")

	go func(ec chan<- error) {
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", c.Addres, c.Port))
		if err != nil {
			log.WithError(err).Error("Failed to listen")
			ec <- err
			return
		}
		if err = gs.Serve(lis); err != nil {
			log.WithError(err).Error("Failed to serve")
			ec <- err
			return
		}
		ec <- nil
	}(ec)
	return gs, ec
}

func (c ServerConfig) httpServerStart() (*http.Server, error) {
	mpObj := mobilpay.CB{}
	var e error
	if mpObj.DBh, e = c.MultiDB.Open(); e != nil {
		return nil, e
	}
	http.HandleFunc("/pay/mobilpayConfirm", mpObj.MobilpayConfirm)
	s := &http.Server{Addr: c.HTTPServer.Address}
	log.Println("Http server started on ", c.HTTPServer.Address)
	go func() { s.ListenAndServe() }()
	return s, nil
}
