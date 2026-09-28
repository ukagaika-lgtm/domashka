package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"expense-server/internal/models"
)

var expenses []models.Expense

var templates = template.Must(
	template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/list.html",
		"web/templates/form.html",
	),
)

func ExpensesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if r.URL.Path == "/expenses/new" {
			err := templates.ExecuteTemplate(w, "layout.html", expenses)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		err := templates.ExecuteTemplate(w, "layout.html", expenses)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		r.ParseForm()

		amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
		if err != nil {
			http.Error(w, "Неверная сумма", http.StatusBadRequest)
			return
		}

		date, err := time.Parse("2006-01-02", r.FormValue("date"))
		if err != nil {
			http.Error(w, "Неверная дата", http.StatusBadRequest)
			return
		}

		expense := models.Expense{
			Amount:      amount,
			Description: r.FormValue("description"),
			Date:        date,
		}

		expenses = append(expenses, expense)

		http.Redirect(w, r, "/expenses", http.StatusSeeOther)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
