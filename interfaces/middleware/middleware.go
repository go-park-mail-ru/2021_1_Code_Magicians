package middleware

import (
	"context"
	"log"
	"net/http"
	"pinterest/application"
	"pinterest/domain/entity"

	"github.com/aws/aws-sdk-go/aws/session"
)

func AuthMid(next http.HandlerFunc, cookieApp application.CookieAppInterface) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, found := CheckCookies(r, cookieApp)
		if !found {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "cookieInfo", cookie)
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	})
}

func NoAuthMid(next http.HandlerFunc, cookieApp application.CookieAppInterface) http.HandlerFunc {
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

//AWS Mid adds AWS session object and bucket name to request context
func AWSMid(next http.HandlerFunc, sess *session.Session, s3BucketName string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "sess", sess)
		ctx = context.WithValue(ctx, "s3BucketName", s3BucketName)
		r = r.Clone(ctx)

		next.ServeHTTP(w, r)
	})
}

var sessions entity.SessionMap = entity.SessionMap{Sessions: make(map[string]entity.CookieInfo)}

// CheckCookies returns *CookieInfo and true if cookie is present in sessions slice, nil and false othervise
func CheckCookies(r *http.Request, cookieApp application.CookieAppInterface) (*entity.CookieInfo, bool) {
	cookie, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		return nil, false
	}

	return cookieApp.CheckCookie(cookie)
}

// JsonContentTypeMid adds "Content-type: application/json" to headers
func JsonContentTypeMid(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
