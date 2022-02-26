package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/DrGermanius/Shortener/internal/store/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/auth"
	"github.com/DrGermanius/Shortener/internal/app/config"
	"github.com/DrGermanius/Shortener/internal/app/models"
)

const (
	gitLink    = "https://github.com"
	yandexLink = "https://yandex.ru"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
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
					ShortURL: app.FullLink(app.ShortLink([]byte(gitLink)))},
				{CorrelationID: "2",
					ShortURL: app.FullLink(app.ShortLink([]byte(yandexLink)))},
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

func TestDeleteLinks(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		link      string
		shortLink string
		want      want
	}{
		{
			link:   yandexLink,
			name:   "positive test #9",
			method: http.MethodDelete,
			want: want{
				code: http.StatusAccepted,
			},
		},
	}
	for _, tt := range tests {
		initTestData()
		t.Run(tt.name, func(t *testing.T) {
			authCookieValue, err := auth.GetSignature()
			require.NoError(t, err)
			authCookie := &http.Cookie{Name: auth.AuthCookie, Value: authCookieValue}

			request := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.link))
			request.AddCookie(authCookie)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(H.AddShortLinkHandler)
			h.ServeHTTP(w, request)

			link := app.ShortLink([]byte(yandexLink))
			req, err := json.Marshal([]string{
				link,
			})
			require.NoError(t, err)

			request = httptest.NewRequest(tt.method, "/api/user/urls", bytes.NewBuffer(req))
			request.AddCookie(authCookie)

			w = httptest.NewRecorder()
			h = H.DeleteLinksHandler
			h.ServeHTTP(w, request)

			res := w.Result()
			require.Equal(t, tt.want.code, res.StatusCode)
			res.Body.Close()
			time.Sleep(time.Second * 2)
			request = httptest.NewRequest(http.MethodGet, "/"+link, nil)
			request.AddCookie(authCookie)

			w = httptest.NewRecorder()
			h = H.GetShortLinkHandler
			h.ServeHTTP(w, request)

			res = w.Result()
			defer res.Body.Close()
			require.Equal(t, http.StatusGone, res.StatusCode)
		})
	}
}

func BenchmarkAddGet(b *testing.B) {
	initTestData()
	authCookieValue, err := auth.GetSignature()
	if err != nil {
		b.Fatal(err)
	}
	authCookie := &http.Cookie{Name: auth.AuthCookie, Value: authCookieValue}
	w := httptest.NewRecorder()
	hAdd := http.HandlerFunc(H.AddShortLinkHandler)
	hGet := http.HandlerFunc(H.GetShortLinkHandler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		s := randStringRunes(10)
		addRequest := httptest.NewRequest("POST", "/", strings.NewReader(s))
		addRequest.AddCookie(authCookie)
		getRequest := httptest.NewRequest("GET", "/", strings.NewReader(s))
		getRequest.AddCookie(authCookie)
		b.StartTimer()

		hAdd.ServeHTTP(w, addRequest)
		hGet.ServeHTTP(w, getRequest)
	}
}

func initTestData() {
	config.SetTestConfig()

	zapl, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer zapl.Sync()
	logger := zapl.Sugar()

	linksMemoryStore, err := memory.NewLinkMemoryStore()
	if err != nil {
		logger.Fatalf("tests init error: %v", err)
	}

	ctx := context.Background()
	wp := app.NewWorkerPool(ctx, logger)
	H = NewHandlers(linksMemoryStore, wp, logger, ctx)

	err = memory.Clear()
	if err != nil {
		logger.Fatalf("tests init error: %v", err)
	}

	_, err = linksMemoryStore.Write(context.Background(), "", gitLink)
	if err != nil {
		logger.Fatalf("tests init error: %v", err)
	}
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type want struct {
	code     int
	response string
	err      error
}
