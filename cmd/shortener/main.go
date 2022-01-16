package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/handlers"
	ml "github.com/DrGermanius/Shortener/internal/app/middlewares"
	"github.com/DrGermanius/Shortener/internal/store"
)

func main() {
	var err error

	c := config.NewConfig()

	storager, err := store.New(c.ConnectionString)
	if err != nil {
		log.Fatalln(err)
	}

	h := handlers.NewHandlers(storager)

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

	log.Println("API started on " + c.ServerAddress)
	log.Fatalln(http.ListenAndServe(c.ServerAddress, r))
}
