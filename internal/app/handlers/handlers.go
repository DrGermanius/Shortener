package handlers

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/store"
)

const (
	Port = "8080"

	host = "http://localhost:" + Port
)

func ShortenerHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	switch req.Method {
	case http.MethodGet:
		s := req.URL.Path[1:] // skip "/" from path

		l, exist := store.LinksMap[s]
		if exist {
			w.Header().Add("Location", l)
			w.WriteHeader(http.StatusTemporaryRedirect)

			_, err := w.Write([]byte{})
			if err != nil {
				log.Print(err)
			}
		} else {
			http.Error(w, app.ErrLinkNotFound.Error(), http.StatusBadRequest)
		}

	case http.MethodPost:
		w.WriteHeader(http.StatusCreated)
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		s := app.ShortLink(b)
		store.LinksMap[s] = string(b)

		full := host + "/" + s

		_, err = w.Write([]byte(full))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

	default:
		http.Error(w, app.ErrMethodNotAllowed.Error(), http.StatusBadRequest)
	}
}
