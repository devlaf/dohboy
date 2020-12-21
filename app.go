package main

import (
	"flag"
	"log"
	"os"

	"dohboy/server"
	"dohboy/test-client"
)

func main() {
	var operation string
	var configPath string
	flag.StringVar(&operation, "op", "server", "operation: [sever|test]")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	if configPath == "" {
		if _, err := os.Stat("./config.yml"); err == nil {
			configPath = "./config.yml"
		}
	}

	if operation == "server" {
		dohboy.Run(configPath)
	} else if operation == "test" {
		test.Run()
	} else {
		log.Fatalf("unknown operation: %v", operation)
	}
}
