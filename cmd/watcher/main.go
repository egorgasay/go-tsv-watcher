package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog"
	"go-tsv-watcher/config"
	"go-tsv-watcher/internal/handler"
	"go-tsv-watcher/internal/storage"
	"go-tsv-watcher/internal/storage/queries"
	"go-tsv-watcher/internal/usecase"
	"go-tsv-watcher/pkg/logger"
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

	loggerInstance := httplog.NewLogger("watcher", httplog.Options{
		Concise: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logic := usecase.New(st, cfg.DirectoryOut, logger.New(loggerInstance))

	go func() {
		err := logic.Process(ctx, cfg.Refresh, cfg.Directory)
		if err != nil {
			log.Fatal(err)
		}
	}()

	h := handler.New(logic)

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
