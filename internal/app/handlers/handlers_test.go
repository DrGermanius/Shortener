package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/DrGermanius/Shortener/internal/app/util"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/auth"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/memory"
	"github.com/DrGermanius/Shortener/internal/app/models"
)

const (
	gitLink    = "https://github.com"
	yandexLink = "https://yandex.ru"
)

var H Handlers

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
			h := http.HandlerFunc(H.AddShortLinkHandler)
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
			h := http.HandlerFunc(H.GetShortLinkHandler)
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
			h := http.HandlerFunc(H.ShortenHandler)
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

func TestGetUserUrls(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		link      string
		shortLink string
		want      want
	}{
		{
			name:   "negative test #6",
			method: http.MethodGet,
			want: want{
				code: http.StatusInternalServerError,
				err:  app.ErrInvalidSignature,
			},
		},
	}
	for _, tt := range tests {
		initTestData()

		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/user/urls", nil)
			authCookie := &http.Cookie{Name: auth.AuthCookie, Value: "123"}
			request.AddCookie(authCookie)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(H.GetUserUrlsHandler)
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

func TestGetUserUrlsWithFakeCookie(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		link      string
		shortLink string
		want      want
	}{
		{
			name:   "negative test #7",
			method: http.MethodGet,
			want: want{
				code: http.StatusNoContent,
				err:  app.ErrUserHasNoRecords,
			},
		},
	}
	for _, tt := range tests {
		initTestData()
		t.Run(tt.name, func(t *testing.T) {
			authCookieValue, err := auth.GetSignature()
			require.NoError(t, err)
			request := httptest.NewRequest(tt.method, "/user/urls", nil)
			authCookie := &http.Cookie{Name: auth.AuthCookie, Value: authCookieValue}
			request.AddCookie(authCookie)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(H.GetUserUrlsHandler)
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

func TestButchLinks(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		link      string
		shortLink string
		want      want
	}{
		{
			name:   "positive test #8",
			method: http.MethodPost,
			want: want{
				code: http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		initTestData()
		t.Run(tt.name, func(t *testing.T) {

			req, err := json.Marshal([]models.BatchOriginal{
				{CorrelationID: "1",
					OriginalURL: gitLink},
				{CorrelationID: "2",
					OriginalURL: yandexLink},
			})
			require.NoError(t, err)

			expectedRes := []models.BatchShort{
				{CorrelationID: "1",
					ShortURL: util.FullLink(app.ShortLink([]byte(gitLink)))},
				{CorrelationID: "2",
					ShortURL: util.FullLink(app.ShortLink([]byte(yandexLink)))},
			}

			request := httptest.NewRequest(tt.method, "/user/urls", bytes.NewBuffer(req))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(H.BatchHandler)
			h.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			var actualRes []models.BatchShort
			err = json.Unmarshal(resBody, &actualRes)
			require.NoError(t, err)

			require.Equal(t, tt.want.code, res.StatusCode)
			require.Equal(t, expectedRes, actualRes)
		})
	}
}

func initTestData() {
	_, err := config.TestConfig()
	if err != nil {
		log.Fatalln(err)
	}

	linksMemoryStore, err := memory.NewLinkMemoryStore()
	if err != nil {
		log.Fatalln(err)
	}

	H = *NewHandlers(linksMemoryStore)

	err = memory.Clear()
	if err != nil {
		log.Fatalln(err)
	}

	_, err = linksMemoryStore.Write(context.Background(), "", gitLink)
	if err != nil {
		log.Fatalln(err)
	}
}

type want struct {
	code     int
	response string
	err      error
}
