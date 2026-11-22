package repository

import (
	"database/sql"
	"fmt"

	"zadanie2/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create сохраняет нового пользователя в базе данных.
func (r *UserRepository) Create(user *models.User) error {
	result, err := r.db.Exec(
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		return fmt.Errorf("создание пользователя: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("получение id пользователя: %w", err)
	}

	user.ID = id
	return nil
}

// GetByEmail находит пользователя по электронной почте.
func (r *UserRepository) GetByEmail(email string) (models.User, error) {
	var user models.User

	err := r.db.QueryRow(
		`SELECT id, email, password_hash FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err != nil {
		return models.User{}, fmt.Errorf("поиск пользователя: %w", err)
	}

	return user, nil
}
