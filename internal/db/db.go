package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	sqliteVec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/shanmeiliu/repo-context-compiler/internal/scanner"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	sqliteVec.Auto()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	database, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return database, nil
}

func Migrate(database *sql.DB) error {
	statements := []string{
		`
		CREATE TABLE IF NOT EXISTS files (
			path TEXT PRIMARY KEY,
			language TEXT,
			size_bytes INTEGER,
			sha256 TEXT,
			summary TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS symbols (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_path TEXT NOT NULL,
			symbol_name TEXT NOT NULL,
			kind TEXT NOT NULL,
			inputs TEXT DEFAULT '[]',
			outputs TEXT DEFAULT '[]',
			side_effects TEXT DEFAULT '[]',
			dependencies TEXT DEFAULT '[]',
			risk TEXT DEFAULT 'unknown',
			summary TEXT DEFAULT '',
			FOREIGN KEY(file_path) REFERENCES files(path)
		);
		`,
		`
		CREATE TABLE IF NOT EXISTS embeddings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ref_type TEXT NOT NULL,
			ref_id TEXT NOT NULL,
			model TEXT NOT NULL,
			dimensions INTEGER NOT NULL,
			embedding BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`,
	}

	for _, stmt := range statements {
		if _, err := database.Exec(stmt); err != nil {
			return err
		}
	}

	// Verify sqlite-vec is loaded.
	var version string
	err := database.QueryRow(`SELECT vec_version();`).Scan(&version)
	if err != nil {
		return fmt.Errorf("sqlite-vec not available: %w", err)
	}

	return nil
}

func UpsertFiles(database *sql.DB, files []scanner.FileInfo) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO files (path, language, size_bytes, sha256, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(path) DO UPDATE SET
			language = excluded.language,
			size_bytes = excluded.size_bytes,
			sha256 = excluded.sha256,
			updated_at = CURRENT_TIMESTAMP;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, file := range files {
		if _, err := stmt.Exec(file.Path, file.Language, file.SizeBytes, file.SHA256); err != nil {
			return err
		}
	}

	return tx.Commit()
}
