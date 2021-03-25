package auth

import (
	"context"
	"net/http"
)

func AuthMid(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, found := CheckCookies(r)
		if !found {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", cookie.UserID)
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	})
}

func NoAuthMid(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, found := CheckCookies(r)
		if found {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
