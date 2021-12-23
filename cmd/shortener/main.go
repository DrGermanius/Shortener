package main

import (
	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/handlers"
	"github.com/DrGermanius/Shortener/internal/app/memory"
	ml "github.com/DrGermanius/Shortener/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	c := config.NewConfig()

	linksMemoryStore, err := memory.NewLinkMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

	h := handlers.NewHandlers(linksMemoryStore)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	r.Use(ml.GzipDecompress)

	r.Get("/{id}", h.GetShortLinkHandler)
	r.Get("/user/urls", h.GetUserUrlsHandler)
	r.Get("/ping", h.PingDatabaseHandler)

	r.Post("/", h.AddShortLinkHandler)
	r.Post("/api/shorten", h.ShortenHandler)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, app.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	})

	log.Println("API started on " + c.ServerAddress)
	log.Fatalln(http.ListenAndServe(c.ServerAddress, r))
}
