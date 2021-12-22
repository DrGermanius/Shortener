package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
)

type gzipReaderCloser struct {
	*gzip.Reader
	io.Closer
}

func (g gzipReaderCloser) Close() error {
	return g.Closer.Close()
}

func GzipDecompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			r.Body = gzipReaderCloser{reader, r.Body}
		}
		next.ServeHTTP(w, r)
	})
}
