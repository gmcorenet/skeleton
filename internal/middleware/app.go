package middleware

import (
	"log"
	"net/http"
	"time"
)

type AppMiddleware struct {
	Logger *log.Logger
}

func NewAppMiddleware() *AppMiddleware {
	return &AppMiddleware{
		Logger: log.Default(),
	}
}

func (m *AppMiddleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.Logger.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		m.Logger.Printf("Request completed in %v", time.Since(start))
	})
}

func (m *AppMiddleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
