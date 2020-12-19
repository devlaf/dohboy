package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"donut/donut_server"
)

func run(config donut.Config) {
	onSignalInterrupt := make(chan os.Signal, 1)

	signal.Notify(
		onSignalInterrupt,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT)

	dohServer, err := donut.CreateDOHServer(config)
	if err != nil {
		log.Fatalf("coud not configure server: %v", err)
	}

	go func() {
		if err := dohServer.ListenAndBlock(); err != nil {
			log.Fatalf("doh server error: %v", err)
		}
	}()
	log.Printf("doh server started.")

	<-onSignalInterrupt
	log.Print("shutting down...\n")

	go func() {
		<-onSignalInterrupt
		log.Fatal("okay, fine! killing immediately you impatient bastard...\n")
	}()

	dohServer.Stop()

	defer os.Exit(0)
	return
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.Parse()

	config, err := donut.FetchConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	run(*config)
}
