package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/moapis/shop/mobilpay"
)

func main() {
	c, err := configure(Default)
	if err != nil {
		log.WithError(err).Fatal("configure()")
	}
	opts, err := c.grpcOpts()
	if err != nil {
		log.WithError(err).Fatal("grpcOpts()")
	}

	s, err := c.newShopServer()
	if err != nil {
		log.WithError(err).Fatal("newShopServer")
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	mobilpay.SetMobilpayVars(c.HTTPServer.MobilpayEndpoint, c.Mobilpay.Signature, c.Mobilpay.PrivateKeyFile, c.Mobilpay.CertificateFile, c.Mobilpay.ConfirmURL, c.Mobilpay.ReturnURL)
	httpServer, err := c.httpServerStart()
	mpkeys := &mobilpay.CB{}
	mpkeys.ParseKeys()
	if err != nil {
		log.WithError(err).Fatal("httpServer")
	}
	gs, ec := c.listenAndServe(s, opts...)
	select {
	case sig := <-sc:
		log.WithField("signal", sig).Info("Shutdown")
		gs.GracefulStop()
		httpServer.Shutdown(context.Background())
	case err = <-ec:
		log.WithError(err).Fatal("Shutdown")
	}
}
