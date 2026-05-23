package db

import (
	"database/sql"
)

type FileRecord struct {
	Path     string
	Language string
	Summary  string
}

type SymbolRecord struct {
	FilePath string
	Name     string
	Kind     string
}

type DependencyRecord struct {
	Source string
	Target string
	Type   string
}

func GetFiles(database *sql.DB) ([]FileRecord, error) {
	rows, err := database.Query(`
		SELECT path, language, summary
		FROM files
		ORDER BY path;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileRecord

	for rows.Next() {
		var file FileRecord

		if err := rows.Scan(&file.Path, &file.Language, &file.Summary); err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, rows.Err()
}

func GetSymbols(database *sql.DB) ([]SymbolRecord, error) {
	rows, err := database.Query(`
		SELECT file_path, symbol_name, kind
		FROM symbols
		ORDER BY file_path, symbol_name;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []SymbolRecord

	for rows.Next() {
		var symbol SymbolRecord

		if err := rows.Scan(&symbol.FilePath, &symbol.Name, &symbol.Kind); err != nil {
			return nil, err
		}

		symbols = append(symbols, symbol)
	}

	return symbols, rows.Err()
}

func GetDependencies(database *sql.DB) ([]DependencyRecord, error) {
	rows, err := database.Query(`
		SELECT source, target, type
		FROM dependencies
		ORDER BY source, target;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []DependencyRecord

	for rows.Next() {
		var dep DependencyRecord

		if err := rows.Scan(&dep.Source, &dep.Target, &dep.Type); err != nil {
			return nil, err
		}

		deps = append(deps, dep)
	}

	return deps, rows.Err()
}
