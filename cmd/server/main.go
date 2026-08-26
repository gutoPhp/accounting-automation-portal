package main

import (
	"log"
	"net/http"
	"os"

	"sheepcontabil/internal/app"
)

func main() {
	addr := env("ADDR", ":8080")
	dataFile := env("DATA_FILE", "data/sheep.json")
	server, err := app.New(dataFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SheepContabil ouvindo em %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.Handler()))
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
