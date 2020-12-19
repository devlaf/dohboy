package main

import (
	"flag"

	"donut/donut_server"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.Parse()

	donut.Run(configPath)
}
