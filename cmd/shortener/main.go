package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/handlers"
	ml "github.com/DrGermanius/Shortener/internal/app/middlewares"
	"github.com/DrGermanius/Shortener/internal/store"
)

func main() {
	var err error
	c := config.NewConfig()
	zapl, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer zapl.Sync()
	logger := zapl.Sugar()

	storager, err := store.New(c.ConnectionString, logger)
	if err != nil {
		logger.Fatalf("can't initialize store: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wp := app.NewWorkerPool(ctx, logger)

	h := handlers.NewHandlers(storager, wp, logger, ctx)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	r.Use(ml.GzipDecompress)

	r.Get("/{id}", h.GetShortLinkHandler)
	r.Get("/user/urls", h.GetUserUrlsHandler)
	r.Get("/ping", h.PingDatabaseHandler)

	r.Post("/", h.AddShortLinkHandler)
	r.Post("/api/shorten", h.ShortenHandler)
	r.Post("/api/shorten/batch", h.BatchHandler)

	r.Delete("/api/user/urls", h.DeleteLinksHandler)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, app.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	})

	logger.Infof("API started on %s", c.ServerAddress)
	go logger.Fatal(http.ListenAndServe(c.ServerAddress, r))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down service...")
}
