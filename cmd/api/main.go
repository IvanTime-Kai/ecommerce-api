package main

import (
	"log"

	"github.com/Ivantime-Kai/ecommerce-api/internal/app"
	"github.com/Ivantime-Kai/ecommerce-api/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	a.Run()
}
