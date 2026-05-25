// Package postgres содержит код для подключения к PostgreSQL и работы с SQL-миграциями.
package postgres

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/gef3dx/api_workinghub/internal/config"
)

// ApplyMigrations применяет все новые SQL-файлы из каталога миграций по порядку.
func ApplyMigrations(db *gorm.DB, cfg config.MigrationsConfig) error {
	if !cfg.AutoApply {
		return nil
	}

	if cfg.Path == "" {
		return fmt.Errorf("migrations path is empty")
	}

	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	files, err := migrationFiles(cfg.Path)
	if err != nil {
		return err
	}

	for _, file := range files {
		applied, err := migrationApplied(db, file)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		statements := splitSQLStatements(string(sqlBytes))
		if len(statements) == 0 {
			if err := markMigrationApplied(db, file); err != nil {
				return err
			}
			continue
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, stmt := range statements {
				if err := tx.Exec(stmt).Error; err != nil {
					return fmt.Errorf("execute migration %s: %w", file, err)
				}
			}

			return markMigrationApplied(tx, file)
		}); err != nil {
			return err
		}
	}

	return nil
}

// ensureMigrationsTable создаёт служебную таблицу для учёта уже применённых миграций.
func ensureMigrationsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`).Error
}

// migrationFiles обходит каталог миграций и возвращает все SQL-файлы в отсортированном виде.
func migrationFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".sql" {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan migrations: %w", err)
	}

	sort.Strings(files)
	return files, nil
}

// migrationApplied проверяет, была ли конкретная миграция уже применена ранее.
func migrationApplied(db *gorm.DB, version string) (bool, error) {
	var count int64
	err := db.Raw("SELECT COUNT(1) FROM schema_migrations WHERE version = ?", version).Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}

	return count > 0, nil
}

// markMigrationApplied фиксирует факт успешного применения миграции.
func markMigrationApplied(db *gorm.DB, version string) error {
	return db.Exec("INSERT INTO schema_migrations(version) VALUES (?)", version).Error
}

// splitSQLStatements делит содержимое SQL-файла на отдельные выражения по символу ';'.
func splitSQLStatements(content string) []string {
	parts := strings.Split(content, ";")
	statements := make([]string, 0, len(parts))

	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		statements = append(statements, stmt)
	}

	return statements
}
