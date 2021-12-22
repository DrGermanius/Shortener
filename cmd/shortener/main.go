package main

import (
	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/handlers"
	ml "github.com/DrGermanius/Shortener/internal/app/middlewares"
	"github.com/DrGermanius/Shortener/internal/app/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	c := config.NewConfig()

	err := store.NewLinksMap()
	if err != nil {
		log.Fatalln(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	r.Use(ml.GzipDecompress)

	r.Get("/{id}", handlers.GetShortLinkHandler)
	r.Get("/user/urls", handlers.GetUserUrlsHandler)
	r.Post("/", handlers.AddShortLinkHandler)
	r.Post("/api/shorten", handlers.ShortenHandler)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, app.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	})

	log.Println("API started on " + c.ServerAddress)
	log.Fatalln(http.ListenAndServe(c.ServerAddress, r))
}
