package main

import (
	"flag"
	"log"
	"os"

	"dohboy/server"
	"dohboy/test-client"
)

func getServerConfigPath() string {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	if configPath == "" {
		if _, err := os.Stat("./config.yml"); err == nil {
			configPath = "./config.yml"
		}
	}

	return configPath
}

func main() {
	var operation string
	flag.StringVar(&operation, "op", "server", "operation: [sever|test]")
	flag.Parse()

	if operation == "server" {
		dohboy.Run(getServerConfigPath())
	} else if operation == "test" {
		test.Run()
	} else {
		log.Fatalf("unknown operation: %v", operation)
	}
}
