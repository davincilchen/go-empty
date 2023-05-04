package main

import (
	"go-empty/pkg/app/server"
	"go-empty/pkg/config"
	"log"
)

const confPath = "./config.json"

func main() {

	cfg, err := config.New(confPath)
	if err != nil {
		log.Fatal(err)
	}

	svr := server.New(cfg)
	svr.Serve()

}
