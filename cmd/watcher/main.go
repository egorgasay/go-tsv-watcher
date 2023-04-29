package main

import (
	"go-tsv-watcher/config"
	"go-tsv-watcher/internal/storage"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/internal/usecase"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()
	st := storage.New(cfg.DBConfig)
	if err := st.Prepare(cfg.DBConfig.Type); err != nil {
		log.Fatal(err)
	}

	logic := usecase.New(st)

	go func() {
		err := logic.Process(cfg.Refresh, cfg.Directory)
		if err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	err := queries.Close()
	if err != nil {
		log.Println(err)
	}

	log.Println("Done!")
}
