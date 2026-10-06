package main

import (
	"net/http"

	"expense-server/zadanie2/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/expenses", handlers.ExpensesHandler)
	mux.HandleFunc("/expenses/new", handlers.ExpensesHandler)

	http.ListenAndServe(":8080", mux)
}
