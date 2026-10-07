package main

import (
	"fmt"
	"log"
	"net/http"

	"zadanie2/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	// Маршруты первого этапа
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/about", handlers.AboutHandler)
	mux.HandleFunc("/ping", handlers.PingHandler)

	// Маршруты второго этапа
	mux.HandleFunc("/expenses", handlers.ExpensesHandler)       // GET — список, POST — добавление
	mux.HandleFunc("/expenses/new", handlers.NewExpenseHandler) // GET — форма

	// Статика (style.css)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
