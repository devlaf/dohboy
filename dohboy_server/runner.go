package dohboy

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Run(configPath string) {

	config, err := FetchConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading specified config file: %v", err)
	}

	onSignalInterrupt := make(chan os.Signal, 1)

	signal.Notify(
		onSignalInterrupt,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT)

	dohServer, err := CreateDOHServer(config)
	if err != nil {
		log.Fatalf("coud not configure server: %v", err)
	}

	go func() {
		if err := dohServer.ListenAndBlock(); err != nil {
			log.Fatalf("dohboy server error: %v", err)
		}
	}()
	log.Printf("dohboy server started.")

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
