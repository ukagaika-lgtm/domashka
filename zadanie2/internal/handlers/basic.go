package handlers

import (
	"fmt"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Привет! Это приложение «Учёт личных трат». Список трат: /expenses")
}

func About(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Учебный проект на Go: учёт личных трат с хранением данных в SQLite.")
}

// Ping вручную проверяет метод (так требует этап 1).
func Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprint(w, "pong")
}
