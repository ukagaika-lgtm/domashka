package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"zadanie2/internal/repository"
)

type StatsHandler struct {
	expenseRepo *repository.ExpenseRepository
}

func NewStatsHandler(expenseRepo *repository.ExpenseRepository) *StatsHandler {
	return &StatsHandler{
		expenseRepo: expenseRepo,
	}
}

func (h *StatsHandler) ByCategory(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if _, err := time.Parse("2006-01-02", from); err != nil {
		http.Error(w, "Некорректная начальная дата", http.StatusBadRequest)
		return
	}

	if _, err := time.Parse("2006-01-02", to); err != nil {
		http.Error(w, "Некорректная конечная дата", http.StatusBadRequest)
		return
	}

	if from > to {
		http.Error(w, "Начальная дата не может быть позже конечной", http.StatusBadRequest)
		return
	}

	stats, err := h.expenseRepo.GetStatsByCategory(userID, from, to)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *StatsHandler) ByMonth(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	year := r.URL.Query().Get("year")

	if _, err := time.Parse("2006", year); err != nil {
		http.Error(w, "Некорректный год", http.StatusBadRequest)
		return
	}

	stats, err := h.expenseRepo.GetStatsByMonth(userID, year)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *StatsHandler) Page(w http.ResponseWriter, r *http.Request) {
	_, ok := requireUserID(w, r)
	if !ok {
		return
	}

	tmpl, err := template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/stats.html",
	)
	if err != nil {
		http.Error(w, "Ошибка загрузки страницы", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", map[string]string{
		"Title": "Статистика",
	})
	if err != nil {
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
