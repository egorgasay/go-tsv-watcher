package main

import (
	"go-tsv-watcher/config"
	"go-tsv-watcher/internal/storage"
	"go-tsv-watcher/internal/usecase"
	"log"
)

func main() {
	cfg := config.New()
	logic := usecase.New(storage.New())
	err := logic.Process(cfg.Directory)
	if err != nil {
		log.Fatal(err)
	}
}
