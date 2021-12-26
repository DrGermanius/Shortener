package handlers

import (
	"context"
	"encoding/json"
	"github.com/DrGermanius/Shortener/internal/app/util"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/auth"
	"github.com/DrGermanius/Shortener/internal/app/models"
)

type LinksStorager interface {
	Get(context.Context, string) (string, error)
	GetByUserID(context.Context, string) (*[]models.LinkJSON, error)
	Write(context.Context, string, string) (string, error)
	BatchWrite(context.Context, string, []models.BatchOriginal) ([]string, error)
	Ping(context.Context) bool
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

	l, err := h.store.Get(req.Context(), s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.Header().Add("Location", l)
	w.WriteHeader(http.StatusTemporaryRedirect)

	_, err = w.Write([]byte{})
	if err != nil {
		log.Print(err)
	}
}

func (h *Handlers) PingDatabaseHandler(w http.ResponseWriter, req *http.Request) {
	if h.store.Ping(context.Background()) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte{})
		if err != nil {
			http.Error(w, "PingDatabaseHandler error", http.StatusInternalServerError)
		}
	}
	http.Error(w, "", http.StatusInternalServerError)
}

func (h *Handlers) GetUserUrlsHandler(w http.ResponseWriter, req *http.Request) {
	uid, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := h.store.GetByUserID(req.Context(), uid)
	if err == app.ErrUserHasNoRecords {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	s, err := h.store.Write(req.Context(), uid, string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	full := util.FullLink(s)

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

	s, err := h.store.Write(req.Context(), uid, sReq.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sRes.Result = util.FullLink(s)
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

func (h *Handlers) BatchHandler(w http.ResponseWriter, req *http.Request) {
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

	var batchReq []models.BatchOriginal

	err = json.Unmarshal(b, &batchReq)
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	shorts, err := h.store.BatchWrite(req.Context(), uid, batchReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	batchRes := make([]models.BatchShort, 0, len(batchReq))
	for i := 0; i < len(batchReq); i++ {
		batchRes = append(batchRes, models.BatchShort{
			CorrelationId: batchReq[i].CorrelationId,
			ShortURL:      util.FullLink(shorts[i]),
		})
	}

	jRes, err := json.Marshal(batchRes)
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
