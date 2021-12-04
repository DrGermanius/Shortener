package handlers

import (
	"encoding/json"
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

	full := config.Config().BaseUrl + s

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(full))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func ShortenHandler(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	sReq := struct {
		URL string `json:"url"`
	}{}

	sRes := struct {
		Result string `json:"result"`
	}{}

	err = json.Unmarshal(b, &sReq)
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	s := app.ShortLink([]byte(sReq.URL))
	store.LinksMap[s] = sReq.URL

	sRes.Result = config.Config().BaseUrl + s
	jRes, _ := json.Marshal(sRes)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(jRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
