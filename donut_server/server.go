package donut

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
	Config     *Config
}

func useTLS(config *Config) bool {
	return config.Server.TLSCertPath != ""
}

func CreateDOHServer(config *Config) (*DOHServer, error) {
	router := createRouter(config)

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
		ReadTimeout:  time.Duration(config.Server.TimeoutMillis.Read) * time.Millisecond,
		WriteTimeout: time.Duration(config.Server.TimeoutMillis.Write) * time.Millisecond,
		IdleTimeout:  time.Duration(config.Server.TimeoutMillis.Idle) * time.Millisecond,
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
	timeout := time.Duration(dohs.Config.Server.TimeoutMillis.Shutdown) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
