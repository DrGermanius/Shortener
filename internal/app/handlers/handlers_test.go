package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DrGermanius/Shortener/internal/app"
	"github.com/DrGermanius/Shortener/internal/app/store"
)

const (
	gitLink    = "https://github.com"
	yandexLink = "https://yandex.ru"
)

func TestHandler(t *testing.T) {
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
			name:      "positive test #2",
			method:    http.MethodGet,
			link:      gitLink,
			shortLink: app.ShortLink([]byte(gitLink)),
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		},
		{
			name:      "negative test #3",
			method:    http.MethodPut,
			link:      gitLink,
			shortLink: app.ShortLink([]byte(gitLink)),
			want: want{
				code: http.StatusBadRequest,
				err:  app.ErrMethodNotAllowed,
			},
		},
		{
			name:      "negative test #4",
			method:    http.MethodDelete,
			link:      gitLink,
			shortLink: app.ShortLink([]byte(gitLink)),
			want: want{
				code: http.StatusBadRequest,
				err:  app.ErrMethodNotAllowed,
			},
		},
		{
			name:      "negative test #5",
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
			var request *http.Request
			if tt.method == http.MethodGet {
				request = httptest.NewRequest(tt.method, "/"+tt.shortLink, nil)
			} else {
				request = httptest.NewRequest(tt.method, "/", strings.NewReader(tt.link))
			}

			w := httptest.NewRecorder()
			h := http.HandlerFunc(ShortenerHandler)
			h.ServeHTTP(w, request)
			res := w.Result()

			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			bodyStr := string(body)

			if tt.want.err == nil {
				if request.Method == http.MethodGet {
					require.Equal(t, tt.want.code, res.StatusCode)
					require.Equal(t, res.Header.Get("Location"), tt.link)
				}
				if request.Method == http.MethodPost {
					require.Equal(t, tt.want.code, res.StatusCode)
					require.Equal(t, tt.want.response, bodyStr)
				}
			} else {
				require.Equal(t, tt.want.code, res.StatusCode)
				require.Error(t, tt.want.err)
			}
		})
	}
}

func initTestData() {
	store.InitLinksMap()

	gitShortLink := app.ShortLink([]byte(gitLink))
	store.LinksMap[gitShortLink] = gitLink
}

type want struct {
	code     int
	response string
	err      error
}
