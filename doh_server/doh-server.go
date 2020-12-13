package doh

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"
)

type DOHServerConfig struct {
	HttpServer *http.Server
	Config     Config
}

type DOHServer interface {
	ListenAndBlock() error
	Stop() error
	RegisterOnStop(callback func())
}

func useTLS(config Config) bool {
	return config.Server.TLSCertPath != ""
}

func CreateDOHServer(config Config) (DOHServer, error) {
	router := createRouter()

	tlsConfig := &tls.Config{}
	if useTLS(config) {
		cert, err := tls.LoadX509KeyPair(config.Server.TLSCertPath, config.Server.TLSKeyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	httpServer := http.Server{
		Addr:         fmt.Sprintf("%v:%v", config.Server.Host, config.Server.Port),
		Handler:      router,
		ReadTimeout:  config.Server.Timeout.Read * time.Second,
		WriteTimeout: config.Server.Timeout.Write * time.Second,
		IdleTimeout:  config.Server.Timeout.Idle * time.Second,
		TLSConfig:    tlsConfig,
	}

	dsc := &DOHServerConfig{
		HttpServer: &httpServer,
		Config:     config,
	}

	return DOHServer(dsc), nil
}

func (dsc *DOHServerConfig) ListenAndBlock() error {
	log.Printf("starting doh server: [%v]", dsc.HttpServer.Addr)

	if useTLS(dsc.Config) {
		if err := dsc.HttpServer.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			return err
		}
	} else {
		if err := dsc.HttpServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
	}

	return nil
}

func (dsc *DOHServerConfig) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), dsc.Config.Server.Timeout.Shutdown*time.Second)
	defer cancel()

	if err := dsc.HttpServer.Shutdown(ctx); err != nil {
		log.Printf("error during http sever shutdown: %v\n", err)
		return err
	}

	log.Printf("doh server stopped.\n")
	return nil
}

func (dsc *DOHServerConfig) RegisterOnStop(callback func()) {
	dsc.HttpServer.RegisterOnShutdown(callback)
}
