package main

import (
	"log"
	"net/http"

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

	h := handlers.NewHandlers(storager, logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	r.Use(ml.GzipDecompress)
	r.Use(ml.CheckAuthCookie)

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
	logger.Fatal(http.ListenAndServe(c.ServerAddress, r))
}
