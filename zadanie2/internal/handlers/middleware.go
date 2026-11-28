package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		log.Printf("%s %s — статус: %d — время: %s",
			r.Method,
			r.URL.Path,
			rw.status,
			time.Since(start),
		)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Паника: %v", err)
				tmpl, templateErr := template.ParseFiles(
					"web/templates/layout.html",
					"web/templates/500.html",
				)
				if templateErr != nil {
					log.Printf("Ошибка шаблона 500: %v", templateErr)
					http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				tmpl.ExecuteTemplate(w, "layout", map[string]string{
					"Title": "Ошибка сервера",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
