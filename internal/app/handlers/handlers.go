package handlers

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/store"
)

func GetShortLinkHandler(w http.ResponseWriter, req *http.Request) {
	s := req.URL.Path[1:] // skip "/" from path; chi.UrlParam not working in tests

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
}

func AddShortLinkHandler(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(b) == 0 {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	s := app.ShortLink(b)
	store.LinksMap[s] = string(b)

	full := config.Config().Full() + "/" + s

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(full))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
