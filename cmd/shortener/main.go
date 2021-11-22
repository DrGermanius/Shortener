package main

import (
	"github.com/DrGermanius/Shortener/internal/app"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/handlers"
	"github.com/DrGermanius/Shortener/internal/app/store"
)

func main() {
	c := config.NewConfig()
	p := strconv.Itoa(c.Port())

	store.NewLinksMap()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{id}", handlers.GetShortLinkHandler)
	r.Post("/", handlers.AddShortLinkHandler)

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, app.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
	})

	log.Println("API started on " + p)
	log.Fatalln(http.ListenAndServe(":"+p, r))
}
