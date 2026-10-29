package repository

import (
	"database/sql"

	"zadanie2/internal/models"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	rows, err := r.db.Query(`
		SELECT id, name
		FROM categories
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		var category models.Category

		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}
