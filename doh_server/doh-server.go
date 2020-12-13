package doh

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"
)

type DOHServer struct {
	HttpServer *http.Server
	Config     Config
}

func useTLS(config Config) bool {
	return config.Server.TLSCertPath != ""
}

func CreateDOHServer(config Config) (*DOHServer, error) {
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

	dohs := &DOHServer{
		HttpServer: &httpServer,
		Config:     config,
	}

	return dohs, nil
}

func (dohs *DOHServer) ListenAndBlock() error {
	log.Printf("starting doh server: [%v]", dohs.HttpServer.Addr)

	if useTLS(dohs.Config) {
		if err := dohs.HttpServer.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			return err
		}
	} else {
		if err := dohs.HttpServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
	}

	return nil
}

func (dohs *DOHServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), dohs.Config.Server.Timeout.Shutdown*time.Second)
	defer cancel()

	if err := dohs.HttpServer.Shutdown(ctx); err != nil {
		log.Printf("error during http sever shutdown: %v\n", err)
		return err
	}

	log.Printf("doh server stopped.\n")
	return nil
}

func (dohs *DOHServer) RegisterOnStop(callback func()) {
	dohs.HttpServer.RegisterOnShutdown(callback)
}
