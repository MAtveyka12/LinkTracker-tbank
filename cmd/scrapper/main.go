package main

import (
	"context"
	"log"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/app"
)

func main() {
	appScrapper, err := app.New()

	if err != nil {
		log.Printf("Failed to create a new app: %s", err.Error())
	}

	err = appScrapper.Start(context.Background())

	if err != nil {
		log.Printf("Failed to start the app: %s", err.Error())
	}
}
