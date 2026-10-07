package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"zadanie2/internal/models"
)

// Хранилище трат в памяти (срез), защищённое мьютексом,
// т.к. обработчики выполняются в разных горутинах.
var (
	mu       sync.Mutex
	expenses []models.Expense
)

// Для каждой страницы — свой набор шаблонов.
// Если парсить list.html и form.html вместе, оба определяют "content",
// и последний файл перезаписывает первый — страницы подменяют друг друга.
var (
	listTmpl = template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/list.html",
	))
	formTmpl = template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/form.html",
	))
)

// ExpensesHandler: GET /expenses — список трат, POST /expenses — добавление.
func ExpensesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		list := make([]models.Expense, len(expenses))
		copy(list, expenses)
		mu.Unlock()

		render(w, listTmpl, list)

	case http.MethodPost:
		addExpense(w, r)

	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// NewExpenseHandler: GET /expenses/new — форма добавления.
func NewExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	render(w, formTmpl, nil)
}

func addExpense(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Не удалось прочитать форму", http.StatusBadRequest)
		return
	}

	// Поддерживаем и точку, и запятую в дробной части
	amountStr := strings.Replace(strings.TrimSpace(r.FormValue("amount")), ",", ".", 1)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Неверная сумма: введите положительное число", http.StatusBadRequest)
		return
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if description == "" {
		http.Error(w, "Описание не может быть пустым", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		http.Error(w, "Неверная дата", http.StatusBadRequest)
		return
	}

	mu.Lock()
	expenses = append(expenses, models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
	})
	mu.Unlock()

	// Post/Redirect/Get: после POST перенаправляем на список
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

func render(w http.ResponseWriter, t *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
