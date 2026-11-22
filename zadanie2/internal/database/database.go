package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite" // регистрирует драйвер с именем "sqlite"
)

// Open открывает файл базы данных SQLite (создаёт его, если файла нет).
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("открытие базы: %w", err)
	}
	// SQLite — файловая база: одно соединение исключает ошибку "database is locked".
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("подключение к базе: %w", err)
	}
	return db, nil
}

// Migrate выполняет все .sql-файлы из папки dir по порядку имён.
func Migrate(db *sql.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("в папке %s нет файлов миграций", dir)
	}

	sort.Strings(files)

	for _, f := range files {
		query, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("чтение %s: %w", f, err)
		}

		_, err = db.Exec(string(query))
		if err != nil {
			// Если колонка уже существует, пропускаем эту миграцию.
			if filepath.Base(f) == "003_add_category_to_expenses.sql" {
				var count int
				checkErr := db.QueryRow(`
					SELECT COUNT(*)
					FROM pragma_table_info('expenses')
					WHERE name = 'category_id'
				`).Scan(&count)

				if checkErr == nil && count > 0 {
					continue
				}
			}

			if filepath.Base(f) == "006_add_user_id_to_expenses.sql" {
				var count int
				checkErr := db.QueryRow(`
				SELECT COUNT(*)
				FROM pragma_table_info('expenses')
				WHERE name = 'user_id'
				`).Scan(&count)
				if checkErr == nil && count > 0 {
					continue
				}
			}

			return fmt.Errorf("миграция %s: %w", f, err)
		}
	}

	return nil
}
