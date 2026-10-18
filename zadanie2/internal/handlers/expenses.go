package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"zadanie2/internal/models"
	"zadanie2/internal/repository"
)

// ExpenseHandler не содержит SQL — только вызовы репозитория.
type ExpenseHandler struct {
	repo     *repository.ExpenseRepository
	listTmpl *template.Template
	formTmpl *template.Template
}

func NewExpenseHandler(repo *repository.ExpenseRepository, tmplDir string) (*ExpenseHandler, error) {
	layout := filepath.Join(tmplDir, "layout.html")
	list, err := template.ParseFiles(layout, filepath.Join(tmplDir, "list.html"))
	if err != nil {
		return nil, err
	}
	form, err := template.ParseFiles(layout, filepath.Join(tmplDir, "form.html"))
	if err != nil {
		return nil, err
	}
	return &ExpenseHandler{repo: repo, listTmpl: list, formTmpl: form}, nil
}

type listPage struct {
	Title    string
	Expenses []models.Expense
}

type formPage struct {
	Title       string
	Action      string
	Amount      string
	Description string
	Date        string
	Error       string
}

// GET /expenses
func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.repo.GetAll()
	if err != nil {
		serverError(w, err)
		return
	}
	render(w, h.listTmpl, http.StatusOK, listPage{Title: "Мои траты", Expenses: expenses})
}

// GET /expenses/new
func (h *ExpenseHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	render(w, h.formTmpl, http.StatusOK, formPage{
		Title:  "Новая трата",
		Action: "/expenses",
		Date:   time.Now().Format("2006-01-02"),
	})
}

// POST /expenses
func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	page := formPage{Title: "Новая трата", Action: "/expenses"}
	e, ok := h.readForm(w, r, &page)
	if !ok {
		return
	}
	if err := h.repo.Create(&e); err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// GET /expenses/{id}/edit
func (h *ExpenseHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	e, err := h.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	render(w, h.formTmpl, http.StatusOK, formPage{
		Title:       "Редактирование траты",
		Action:      fmt.Sprintf("/expenses/%d/edit", id),
		Amount:      strconv.FormatFloat(e.Amount, 'f', 2, 64),
		Description: e.Description,
		Date:        e.Date.Format("2006-01-02"),
	})
}

// POST /expenses/{id}/edit
func (h *ExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	page := formPage{Title: "Редактирование траты", Action: fmt.Sprintf("/expenses/%d/edit", id)}
	e, ok := h.readForm(w, r, &page)
	if !ok {
		return
	}
	e.ID = id
	err := h.repo.Update(e)
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// POST /expenses/{id}/delete
func (h *ExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	err := h.repo.Delete(id)
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, "/expenses", http.StatusSeeOther)
}

// readForm разбирает и проверяет форму. При ошибке сам показывает форму заново.
func (h *ExpenseHandler) readForm(w http.ResponseWriter, r *http.Request, page *formPage) (models.Expense, bool) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Не удалось прочитать форму", http.StatusBadRequest)
		return models.Expense{}, false
	}
	page.Amount = strings.TrimSpace(r.FormValue("amount"))
	page.Description = strings.TrimSpace(r.FormValue("description"))
	page.Date = strings.TrimSpace(r.FormValue("date"))

	e, msg := validate(page.Amount, page.Description, page.Date)
	if msg != "" {
		page.Error = msg
		render(w, h.formTmpl, http.StatusBadRequest, page)
		return models.Expense{}, false
	}
	return e, true
}

func validate(amountStr, description, dateStr string) (models.Expense, string) {
	amount, err := strconv.ParseFloat(strings.ReplaceAll(amountStr, ",", "."), 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return models.Expense{}, "Сумма должна быть положительным числом, например 250 или 99,90"
	}
	if description == "" {
		return models.Expense{}, "Введите описание траты"
	}
	if utf8.RuneCountInString(description) > 200 {
		return models.Expense{}, "Описание должно быть не длиннее 200 символов"
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return models.Expense{}, "Укажите корректную дату"
	}
	return models.Expense{
		Amount:      math.Round(amount*100) / 100,
		Description: description,
		Date:        date,
	}, ""
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return 0, false
	}
	return id, true
}

// render сначала рендерит в буфер, чтобы при ошибке шаблона не отдать полстраницы.
func render(w http.ResponseWriter, t *template.Template, status int, data any) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("ошибка: %v", err)
	http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
}
