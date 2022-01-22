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
	"github.com/DrGermanius/Shortener/internal/app/models"
	"github.com/DrGermanius/Shortener/internal/store"
)

type Handlers struct {
	store  store.LinksStorager
	logger *zap.SugaredLogger
}

func NewHandlers(store store.LinksStorager, logger *zap.SugaredLogger) Handlers {
	return Handlers{store: store, logger: logger}
}

func (h Handlers) GetShortLinkHandler(w http.ResponseWriter, req *http.Request) {
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

func (h Handlers) GetUserUrlsHandler(w http.ResponseWriter, req *http.Request) {
	uid, _ := req.Context().Value("uid").(string)

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

func (h Handlers) AddShortLinkHandler(w http.ResponseWriter, req *http.Request) {
	uid, _ := req.Context().Value("uid").(string)

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

func (h Handlers) ShortenHandler(w http.ResponseWriter, req *http.Request) {
	uid, _ := req.Context().Value("uid").(string)

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

func (h Handlers) BatchHandler(w http.ResponseWriter, req *http.Request) {
	uid, _ := req.Context().Value("uid").(string)

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

func (h Handlers) DeleteLinksHandler(w http.ResponseWriter, req *http.Request) {
	uid, _ := req.Context().Value("uid").(string)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	time.AfterFunc(time.Second*20, cancel)

	wp := app.NewDeleteWorkerPool(ctx, uid, links, h.store.Delete, h.logger)
	go wp.Run()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	_, err = w.Write([]byte{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
