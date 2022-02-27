// Package handlers stores Shortener api handlers.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/auth"
	"github.com/DrGermanius/Shortener/internal/app/models"
	"github.com/DrGermanius/Shortener/internal/store"
)

type Handlers struct {
	store      store.LinksStorager
	workerPool app.WorkerPool
	logger     *zap.SugaredLogger
	context    context.Context
}

func NewHandlers(store store.LinksStorager, wp app.WorkerPool, logger *zap.SugaredLogger, context context.Context) Handlers {
	return Handlers{store: store, workerPool: wp, logger: logger, context: context}
}

// GetShortLinkHandler redirect client to full url address by short representation.
func (h Handlers) GetShortLinkHandler(w http.ResponseWriter, req *http.Request) {
	_, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := req.URL.Path[1:] // skip "/" from path; chi.UrlParam not working in tests

	l, err := h.store.Get(req.Context(), s)
	if err != nil {
		if errors.Is(err, app.ErrDeletedLink) {
			http.Error(w, err.Error(), http.StatusGone)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", l)
	w.WriteHeader(http.StatusTemporaryRedirect)

	_, err = w.Write([]byte{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PingDatabaseHandler checks if links store is available.
func (h Handlers) PingDatabaseHandler(w http.ResponseWriter, req *http.Request) {
	if h.store.Ping(context.Background()) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte{})
		if err != nil {
			http.Error(w, "PingDatabaseHandler error", http.StatusInternalServerError)
		}
	}
	http.Error(w, "", http.StatusInternalServerError)
}

// GetUserUrlsHandler return user's loaded links by userID.
func (h Handlers) GetUserUrlsHandler(w http.ResponseWriter, req *http.Request) {
	uid, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := h.store.GetByUserID(req.Context(), uid)
	if err != nil {
		if errors.Is(err, app.ErrUserHasNoRecords) {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
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

// AddShortLinkHandler create and return short representation of URL address and save it.
func (h Handlers) AddShortLinkHandler(w http.ResponseWriter, req *http.Request) {
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

	linkAlreadyExist := false
	s, err := h.store.Write(req.Context(), uid, string(b))
	if err != nil {
		if errors.Is(err, app.ErrLinkAlreadyExists) {
			linkAlreadyExist = true
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	full := app.FullLink(s)

	if linkAlreadyExist {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	_, err = w.Write([]byte(full))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// ShortenHandler create and return short representation of URL address and save it via JSON.
func (h Handlers) ShortenHandler(w http.ResponseWriter, req *http.Request) {
	uid, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	sReq := models.ShortenRequest{}
	sRes := models.ShortenResponse{}

	err = json.Unmarshal(b, &sReq)
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	linkAlreadyExist := false
	s, err := h.store.Write(req.Context(), uid, sReq.URL)
	if err != nil {
		if errors.Is(err, app.ErrLinkAlreadyExists) {
			linkAlreadyExist = true
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	sRes.Result = app.FullLink(s)
	jRes, err := json.Marshal(sRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if linkAlreadyExist {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	_, err = w.Write(jRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// BatchHandler takes a couple of URL addresses via JSON, create and return short representation of that and save it.
func (h Handlers) BatchHandler(w http.ResponseWriter, req *http.Request) {
	uid, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

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
			CorrelationID: batchReq[i].CorrelationID,
			ShortURL:      app.FullLink(shorts[i]),
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

// DeleteLinksHandler takes a couple of user's URL addresses via JSON and delete from store.
func (h Handlers) DeleteLinksHandler(w http.ResponseWriter, req *http.Request) {
	uid, err := checkAuthCookie(w, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}

	var links []string
	err = json.Unmarshal(b, &links)
	if err != nil {
		http.Error(w, app.ErrEmptyBodyPostReq.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(h.context, time.Second*20)
	time.AfterFunc(time.Second*20, cancel)
	go h.workerPool.StartDeleteWorker(uid, links, h.store.Delete, ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	_, err = w.Write([]byte{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// checkAuthCookie set or validate user cookie and authenticate user.
func checkAuthCookie(w http.ResponseWriter, req *http.Request) (string, error) {
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
		return uid, nil
	}

	uid, err = auth.CheckSignature(authCookie.Value)
	if err != nil {
		return "", err
	}
	return uid, nil
}
