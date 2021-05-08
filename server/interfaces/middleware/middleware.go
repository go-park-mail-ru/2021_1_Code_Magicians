package middleware

import (
	"context"
	"log"
	"net/http"

	"pinterest/usage"
	"pinterest/domain/entity"

	"github.com/gorilla/csrf"
)

func AuthMid(next http.HandlerFunc, cookieApp usage.AuthAppInterface) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, found := CheckCookies(r, cookieApp)
		if !found {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), entity.CookieInfoKey, cookie)
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	})
}

func NoAuthMid(next http.HandlerFunc, cookieApp usage.AuthAppInterface) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, found := CheckCookies(r, cookieApp)
		if found {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// PanicMid logges error if handler errors
func PanicMid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CSRFSettingMid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r != nil {
			if r.Header.Get("X-CSRF-Token") == "" {
				token := csrf.Token(r)
				w.Header().Set("X-CSRF-Token", token)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// CheckCookies returns *CookieInfo and true if cookie is present in sessions slice, nil and false othervise
func CheckCookies(r *http.Request, cookieApp usage.AuthAppInterface) (*entity.CookieInfo, bool) {
	cookie, err := r.Cookie(entity.CookieNameKey)
	if err == http.ErrNoCookie {
		return nil, false
	}

	return cookieApp.CheckCookie(cookie)
}
