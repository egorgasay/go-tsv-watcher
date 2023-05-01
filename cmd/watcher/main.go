package main

import (
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
	cfg := config.New()
	st, err := storage.New(cfg.DBConfig)
	if err != nil {
		log.Fatal(err)
	}

	logic := usecase.New(st, "out") // TODO: use config

	go func() {
		err := logic.Process(cfg.Refresh, cfg.Directory)
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

	err = queries.Close()
	if err != nil {
		log.Println(err)
	}

	log.Println("Done!")
}
