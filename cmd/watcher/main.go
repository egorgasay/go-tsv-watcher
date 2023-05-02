package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go-tsv-watcher/config"
	resthandler "go-tsv-watcher/internal/handler/rest"
	"go-tsv-watcher/internal/storage"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	st, err := storage.New(cfg.DBConfig)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logic := usecase.New(st, cfg.DirectoryOut)

	go func() {
		err := logic.Process(ctx, cfg.Refresh, cfg.Directory)
		if err != nil {
			log.Fatal(err)
		}
	}()

	h := resthandler.New(logic)

	router := chi.NewRouter()
	router.Group(h.PublicRoutes)

	// http server
	go func() {
		log.Println("HTTP server started on ", cfg.HTTP)
		if err := http.ListenAndServe(cfg.HTTP, router); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	cancel()

	err = queries.Close()
	if err != nil {
		log.Println(err)
	}

	log.Println("Done!")
}
