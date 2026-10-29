package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"zadanie2/internal/models"
)

// ErrNotFound возвращается, если траты с таким id нет.
var ErrNotFound = errors.New("трата не найдена")

const dateLayout = "2006-01-02"

// ExpenseRepository содержит весь SQL-код для работы с тратами.
type ExpenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) Create(e *models.Expense) error {
	res, err := r.db.Exec(
		`INSERT INTO expenses (amount, description, date, category_id) VALUES (?, ?, ?, ?)`,
		e.Amount, e.Description, e.Date.Format(dateLayout), e.CategoryID,
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

func (r *ExpenseRepository) GetAll() ([]models.Expense, error) {
	rows, err := r.db.Query(`
		SELECT e.id, e.amount, e.description, e.date,
		       e.category_id, c.name
		FROM expenses e
		JOIN categories c ON c.id = e.category_id
		ORDER BY e.date DESC, e.id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("получение трат: %w", err)
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		var date string

		err := rows.Scan(
			&e.ID,
			&e.Amount,
			&e.Description,
			&date,
			&e.CategoryID,
			&e.CategoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("чтение траты: %w", err)
		}

		e.Date, err = time.Parse("2006-01-02", date)
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

func (r *ExpenseRepository) GetByID(id int64) (models.Expense, error) {
	row := r.db.QueryRow(
		`SELECT id, amount, description, date, category_id FROM expenses WHERE id = ?`, id)
	e, err := scanExpense(row)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Expense{}, ErrNotFound
	}
	return e, err
}

func (r *ExpenseRepository) Update(e models.Expense) error {
	res, err := r.db.Exec(
		`UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ?`,
		e.Amount, e.Description, e.Date.Format(dateLayout), e.CategoryID, e.ID,
	)
	if err != nil {
		return fmt.Errorf("обновление траты: %w", err)
	}
	return checkAffected(res)
}

func (r *ExpenseRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM expenses WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("удаление траты: %w", err)
	}
	return checkAffected(res)
}

// scanner — общее у *sql.Row и *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanExpense(s scanner) (models.Expense, error) {
	var e models.Expense
	var date string
	if err := s.Scan(&e.ID, &e.Amount, &e.Description, &date, &e.CategoryID); err != nil {
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
	categoryID int64,
	from string,
	to string,
) ([]models.Expense, error) {
	query := `
		SELECT e.id, e.amount, e.description, e.date,
		       e.category_id, c.name
		FROM expenses e
		JOIN categories c ON c.id = e.category_id
		WHERE 1 = 1
	`
	var args []any

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
			&e.Amount,
			&e.Description,
			&date,
			&e.CategoryID,
			&e.CategoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("чтение траты: %w", err)
		}

		e.Date, err = time.Parse("2006-01-02", date)
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
