package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/404.html",
	)
	if err != nil {
		log.Printf("Ошибка шаблона 404: %v", err)
		http.Error(w, "Страница не найдена", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	tmpl.ExecuteTemplate(w, "layout", map[string]string{
		"Title": "Страница не найдена",
	})
}
