package main

import (
	"log"
	"net/http"

	"zadanie2/internal/database"
	"zadanie2/internal/handlers"
	"zadanie2/internal/repository"
)

func main() {
	db, err := database.Open("expenses.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.Migrate(db, "migrations"); err != nil {
		log.Fatal(err)
	}

	repo := repository.NewExpenseRepository(db)
	expenses, err := handlers.NewExpenseHandler(repo, "web/templates")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	// Этап 1
	mux.HandleFunc("GET /{$}", handlers.Home)
	mux.HandleFunc("GET /about", handlers.About)
	mux.HandleFunc("/ping", handlers.Ping)

	// Статика
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Траты (CRUD)
	mux.HandleFunc("GET /expenses", expenses.List)
	mux.HandleFunc("GET /expenses/new", expenses.NewForm)
	mux.HandleFunc("POST /expenses", expenses.Create)
	mux.HandleFunc("GET /expenses/{id}/edit", expenses.EditForm)
	mux.HandleFunc("POST /expenses/{id}/edit", expenses.Update)
	mux.HandleFunc("POST /expenses/{id}/delete", expenses.Delete)

	log.Println("Сервер запущен: http://localhost:8080/expenses")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
