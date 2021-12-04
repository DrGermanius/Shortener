package main

import (
	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"

	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/handlers"
	"github.com/DrGermanius/Shortener/internal/app/store"
)

func main() {
	c := config.NewConfig()

	store.NewLinksMap()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{id}", handlers.GetShortLinkHandler)
	r.Post("/", handlers.AddShortLinkHandler)
	r.Post("/api/shorten", handlers.ShortenHandler)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, app.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	})

	log.Println("API started on " + c.ServerAddress)
	log.Fatalln(http.ListenAndServe(c.ServerAddress, r))
}
