package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/auth"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/models"
)

type LinksStorager interface {
	Get(string) (string, bool)
	GetByUserID(id string) []models.LinkJSON
	Write(uuid, long string) (string, error)
	Ping() bool
}

type Handlers struct {
	store LinksStorager
}

func NewHandlers(store LinksStorager) *Handlers {
	return &Handlers{store: store}
}

func (h *Handlers) GetShortLinkHandler(w http.ResponseWriter, req *http.Request) {
	_, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := req.URL.Path[1:] // skip "/" from path; chi.UrlParam not working in tests

	l, exist := h.store.Get(s)
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

func (h *Handlers) PingDatabaseHandler(w http.ResponseWriter, req *http.Request) {
	if h.store.Ping() {
		w.WriteHeader(http.StatusTemporaryRedirect)
		_, err := w.Write([]byte{})
		if err != nil {
			log.Print(err)
		}
	}
	http.Error(w, app.ErrLinkNotFound.Error(), http.StatusInternalServerError)
}

func (h *Handlers) GetUserUrlsHandler(w http.ResponseWriter, req *http.Request) {
	uid, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := h.store.GetByUserID(uid)
	if len(res) == 0 {
		http.Error(w, app.ErrUserHasNoRecords.Error(), http.StatusNoContent)
		return
	}

	jRes, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(jRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *Handlers) AddShortLinkHandler(w http.ResponseWriter, req *http.Request) {
	uid, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	s, err := h.store.Write(uid, string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	full := config.Config().BaseURL + "/" + s

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(full))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (h *Handlers) ShortenHandler(w http.ResponseWriter, req *http.Request) {
	uid, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	sReq := models.ShortenRequest{}
	sRes := models.ShortenResponse{}

	err = json.Unmarshal(b, &sReq)
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	s, err := h.store.Write(uid, sReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sRes.Result = config.Config().BaseURL + "/" + s
	jRes, err := json.Marshal(sRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(jRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func checkAuthCookie(w http.ResponseWriter, req *http.Request) (string, error) { //todo as middleware?
	uid := ""
	authCookie, err := req.Cookie(auth.AuthCookie)
	if err != nil {
		signaturedUUID, err := auth.GetSignature()
		if err != nil {
			return "", err
		}

		uid, err = auth.CheckSignature(signaturedUUID)
		if err != nil {
			return "", err
		}
		http.SetCookie(w, &http.Cookie{Name: auth.AuthCookie, Value: signaturedUUID})

	} else {
		uid, err = auth.CheckSignature(authCookie.Value)
		if err != nil {
			return "", err
		}
	}
	return uid, nil
}
