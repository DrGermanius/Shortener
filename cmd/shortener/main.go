package main

import (
	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/database"
	"github.com/DrGermanius/Shortener/internal/app/handlers"
	"github.com/DrGermanius/Shortener/internal/app/memory"
	ml "github.com/DrGermanius/Shortener/internal/app/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	var err error
	var store handlers.LinksStorager

	c := config.NewConfig()

	if c.ConnectionString == "" {
		store, err = memory.NewLinkMemoryStore()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Service uses inmemory storage")
	} else {
		store, err = database.NewDatabaseStore(c.ConnectionString)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("Service uses database")
	}

	h := handlers.NewHandlers(store)

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

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, app.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	})

	log.Println("API started on " + c.ServerAddress)
	log.Fatalln(http.ListenAndServe(c.ServerAddress, r))
}
