package middlewares

import (
	"context"
	"net/http"

	"github.com/DrGermanius/Shortener/internal/app/auth"
)

func CheckAuthCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := ""
		authCookie, err := r.Cookie(auth.AuthCookie)
		if err != nil {
			signaturedUUID, err := auth.GetSignature()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			uid, err = auth.CheckSignature(signaturedUUID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			http.SetCookie(w, &http.Cookie{Name: auth.AuthCookie, Value: signaturedUUID})
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		uid, err = auth.CheckSignature(authCookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		r = r.WithContext(context.WithValue(r.Context(), "uid", uid))
		next.ServeHTTP(w, r)
	})
}
