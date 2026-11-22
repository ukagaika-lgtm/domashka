package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"zadanie2/internal/models"
)

var ErrNotFound = errors.New("трата не найдена")

const dateLayout = "2006-01-02"

type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(e *models.Expense) error {
	res, err := r.db.Exec(
		`INSERT INTO expenses (user_id, amount, description, date, category_id)
		 VALUES (?, ?, ?, ?, ?)`,
		e.UserID, e.Amount, e.Description, e.Date.Format(dateLayout), e.CategoryID,
	)
	if err != nil {
		return fmt.Errorf("создание траты: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("получение id: %w", err)
	}

	e.ID = id
	return nil
}

func (r *ExpenseRepository) GetAll(userID int64) ([]models.Expense, error) {
	return r.GetFiltered(userID, 0, "", "")
}

func (r *ExpenseRepository) GetByID(id, userID int64) (models.Expense, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, amount, description, date, category_id
		 FROM expenses WHERE id = ? AND user_id = ?`,
		id, userID,
	)

	e, err := scanExpense(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Expense{}, ErrNotFound
	}
	return e, err
}

func (r *ExpenseRepository) Update(e models.Expense) error {
	res, err := r.db.Exec(
		`UPDATE expenses
		 SET amount = ?, description = ?, date = ?, category_id = ?
		 WHERE id = ? AND user_id = ?`,
		e.Amount, e.Description, e.Date.Format(dateLayout),
		e.CategoryID, e.ID, e.UserID,
	)
	if err != nil {
		return fmt.Errorf("обновление траты: %w", err)
	}
	return checkAffected(res)
}

func (r *ExpenseRepository) Delete(id, userID int64) error {
	res, err := r.db.Exec(
		`DELETE FROM expenses WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("удаление траты: %w", err)
	}
	return checkAffected(res)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExpense(s scanner) (models.Expense, error) {
	var e models.Expense
	var date string

	if err := s.Scan(
		&e.ID, &e.UserID, &e.Amount, &e.Description, &date, &e.CategoryID,
	); err != nil {
		return e, fmt.Errorf("чтение траты: %w", err)
	}

	t, err := time.Parse(dateLayout, date)
	if err != nil {
		return e, fmt.Errorf("неверная дата %q в базе: %w", date, err)
	}

	e.Date = t
	return e, nil
}

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("проверка результата: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ExpenseRepository) GetFiltered(
	userID int64,
	categoryID int64,
	from string,
	to string,
) ([]models.Expense, error) {
	query := `
		SELECT e.id, e.user_id, e.amount, e.description, e.date,
		       e.category_id, c.name
		FROM expenses e
		JOIN categories c ON c.id = e.category_id
		WHERE e.user_id = ?
	`

	args := []any{userID}

	if categoryID > 0 {
		query += " AND e.category_id = ?"
		args = append(args, categoryID)
	}

	if from != "" {
		query += " AND e.date >= ?"
		args = append(args, from)
	}

	if to != "" {
		query += " AND e.date <= ?"
		args = append(args, to)
	}

	query += " ORDER BY e.date DESC, e.id DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("получение отфильтрованных трат: %w", err)
	}
	defer rows.Close()

	var expenses []models.Expense

	for rows.Next() {
		var e models.Expense
		var date string

		err := rows.Scan(
			&e.ID,
			&e.UserID,
			&e.Amount,
			&e.Description,
			&date,
			&e.CategoryID,
			&e.CategoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("чтение траты: %w", err)
		}

		e.Date, err = time.Parse(dateLayout, date)
		if err != nil {
			return nil, fmt.Errorf("преобразование даты: %w", err)
		}

		expenses = append(expenses, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("перебор трат: %w", err)
	}

	return expenses, nil
}
