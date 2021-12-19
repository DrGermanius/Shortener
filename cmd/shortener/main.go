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
	"github.com/DrGermanius/Shortener/internal/app/store"
)

func main() {
	c := config.NewConfig()

	err := store.NewLinksMap()
	if err != nil {
		log.Fatalln(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(ml.GzipHandle)

	r.Get("/{id}", handlers.GetShortLinkHandler)
	r.Post("/", handlers.AddShortLinkHandler)
	r.Post("/api/shorten", handlers.ShortenHandler)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, app.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	})

	log.Println("API started on " + c.ServerAddress)
	log.Fatalln(http.ListenAndServe(c.ServerAddress, r))
}
