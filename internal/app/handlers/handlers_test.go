package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/models"
	"github.com/DrGermanius/Shortener/internal/app/store"
)

const (
	gitLink    = "https://github.com"
	yandexLink = "https://yandex.ru"
)

func TestPostHandler(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		link      string
		shortLink string
		want      want
	}{
		{
			name:   "positive test #1",
			method: http.MethodPost,
			link:   yandexLink,
			want: want{
				code:     http.StatusCreated,
				response: "http://localhost:8080/" + app.ShortLink([]byte(yandexLink)),
			},
		},
		{
			name:   "negative test #2",
			method: http.MethodPost,
			link:   "",
			want: want{
				code: http.StatusBadRequest,
				err:  app.ErrEmptyBodyPostReq,
			},
		},
	}
	for _, tt := range tests {
		initTestData()

		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.link))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(AddShortLinkHandler)
			h.ServeHTTP(w, request)
			res := w.Result()

			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			bodyStr := string(body)

			if tt.want.err != nil {
				assert.Equal(t, tt.want.code, res.StatusCode)
				require.Error(t, tt.want.err)
				return
			}

			assert.Equal(t, tt.want.code, res.StatusCode)
			require.Equal(t, tt.want.response, bodyStr)

		})
	}
}

func TestGetHandler(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		link      string
		shortLink string
		want      want
	}{
		{
			name:      "positive test #3",
			method:    http.MethodGet,
			link:      gitLink,
			shortLink: app.ShortLink([]byte(gitLink)),
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		},
		{
			name:      "negative test #4",
			method:    http.MethodGet,
			link:      yandexLink,
			shortLink: app.ShortLink([]byte(yandexLink)),
			want: want{
				code: http.StatusBadRequest,
				err:  app.ErrLinkNotFound,
			},
		},
	}
	for _, tt := range tests {
		initTestData()

		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/"+tt.shortLink, nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetShortLinkHandler)
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			if tt.want.err != nil {
				assert.Equal(t, tt.want.code, res.StatusCode)
				require.Error(t, tt.want.err)
				return
			}

			require.Equal(t, tt.want.code, res.StatusCode)
			require.Equal(t, res.Header.Get("Location"), tt.link)

		})
	}
}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		link      string
		shortLink string
		want      want
	}{
		{
			name:      "positive test #5",
			method:    http.MethodPost,
			link:      gitLink,
			shortLink: "http://localhost:8080/" + app.ShortLink([]byte(gitLink)),
			want: want{
				code: http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		initTestData()

		sReq := models.ShortenRequest{URL: tt.link}
		sRes := models.ShortenResponse{}

		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(sReq)
			require.NoError(t, err)

			request := httptest.NewRequest(tt.method, "/api/shorten", bytes.NewBuffer(body))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(ShortenHandler)
			h.ServeHTTP(w, request)
			res := w.Result()

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			err = json.Unmarshal(resBody, &sRes)
			require.NoError(t, err)

			if tt.want.err != nil {
				assert.Equal(t, tt.want.code, res.StatusCode)
				require.Error(t, tt.want.err)
				return
			}

			require.Equal(t, sRes.Result, tt.shortLink)
			assert.Equal(t, res.Header.Get("Content-Type"), "application/json")

		})
	}
}

func initTestData() {
	config.Suite()

	err := store.NewLinksMap()
	if err != nil {
		log.Fatalln(err)
	}

	err = store.Clear()
	if err != nil {
		log.Fatalln(err)
	}

	gitShortLink := app.ShortLink([]byte(gitLink))
	store.LinksMap[gitShortLink] = gitLink
}

type want struct {
	code     int
	response string
	err      error
}
