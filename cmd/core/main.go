package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gerladeno/homie-core/internal"
	"github.com/gerladeno/homie-core/internal/rest"
	"github.com/gerladeno/homie-core/internal/storage"
	"github.com/gerladeno/homie-core/pkg/logging"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/sirupsen/logrus"
)

const httpPort = 3001

//go:embed public.pub
var publicSigningKey []byte

var (
	version = `0.0.0`
	pgDSN   = os.Getenv("PG_DSN")
	domain  = os.Getenv("APP_DOMAIN")
)

func main() {
	log := logging.GetLogger(true)
	ctx := context.Background()
	store, err := storage.New(ctx, log, pgDSN)
	if err != nil {
		log.Panicf("err initing pg: %v", err)
	}
	if err = store.Migrate(); err != nil {
		log.Panicf("err migrating pg: %v", err)
	}
	app := internal.NewApp(log, store)
	router := rest.NewRouter(log, app, mustGetPublicKey(publicSigningKey), domain, version)
	if err = startServer(ctx, router, log); err != nil {
		log.Panic(err)
	}
}

func startServer(ctx context.Context, router http.Handler, log *logrus.Logger) error {
	log.Infof("starting server on port %d", httpPort)
	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", httpPort),
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		Handler:           router,
	}
	errCh := make(chan error)
	go func() {
		if err := s.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	select {
	case err := <-errCh:
		return err
	case <-sigCh:
	}
	log.Info("terminating...")
	gfCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.Shutdown(gfCtx)
}

func mustGetPublicKey(keyBytes []byte) *rsa.PublicKey {
	if len(keyBytes) == 0 {
		panic("file public.pub is missing or invalid")
	}
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic("unable to decode public key to blocks")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return key.(*rsa.PublicKey)
}
