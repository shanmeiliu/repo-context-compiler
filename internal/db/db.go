package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	sqliteVec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/shanmeiliu/repo-context-compiler/internal/parser"
	"github.com/shanmeiliu/repo-context-compiler/internal/scanner"

	_ "github.com/mattn/go-sqlite3"

	repoctx "github.com/shanmeiliu/repo-context-compiler/internal"
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
		CREATE TABLE IF NOT EXISTS dependencies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source TEXT NOT NULL,
		target TEXT NOT NULL,
		type TEXT NOT NULL
		);
		`,
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

func UpsertDependencies(
	database *sql.DB,
	deps []parser.Dependency,
) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO dependencies (
			source,
			target,
			type
		)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, dep := range deps {
		_, err := stmt.Exec(
			dep.Source,
			dep.Target,
			dep.Type,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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

func UpsertSymbols(database *sql.DB, symbols []parser.Symbol) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO symbols (
			file_path,
			symbol_name,
			kind,
			summary
		)
		VALUES (?, ?, ?, '')
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, symbol := range symbols {
		_, err := stmt.Exec(
			symbol.FilePath,
			symbol.Name,
			symbol.Kind,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func UpsertFileSummaries(database *sql.DB, summaries []repoctx.FileSummary) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE files
		SET summary = ?, updated_at = CURRENT_TIMESTAMP
		WHERE path = ?;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, summary := range summaries {
		if _, err := stmt.Exec(summary.Summary, summary.Path); err != nil {
			return err
		}
	}

	return tx.Commit()
}

type FileSummaryState struct {
	Path       string
	SHA256     string
	HasSummary bool
}

func GetFileSummaryState(database *sql.DB) (map[string]FileSummaryState, error) {
	rows, err := database.Query(`
		SELECT path, sha256, summary
		FROM files;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]FileSummaryState{}

	for rows.Next() {
		var path string
		var sha string
		var summary string

		if err := rows.Scan(&path, &sha, &summary); err != nil {
			return nil, err
		}

		result[path] = FileSummaryState{
			Path:       path,
			SHA256:     sha,
			HasSummary: summary != "",
		}
	}

	return result, rows.Err()
}

func GetFileSummaries(database *sql.DB) (map[string]string, error) {
	rows, err := database.Query(`
		SELECT path, summary
		FROM files
		WHERE summary IS NOT NULL AND summary != '';
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]string{}

	for rows.Next() {
		var path string
		var summary string

		if err := rows.Scan(&path, &summary); err != nil {
			return nil, err
		}

		result[path] = summary
	}

	return result, rows.Err()
}
