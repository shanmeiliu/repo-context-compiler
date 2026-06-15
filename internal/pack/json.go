package pack

import (
	"encoding/json"
	"os"
	"sort"
	"time"

	"github.com/shanmeiliu/repo-context-compiler/internal/parser"
	"github.com/shanmeiliu/repo-context-compiler/internal/scanner"
)

type ContextPack struct {
	SchemaVersion string              `json:"schema_version"`
	GeneratedAt   string              `json:"generated_at"`
	Files         []scanner.FileInfo  `json:"files"`
	Symbols       []parser.Symbol     `json:"symbols"`
	Summaries     map[string]string   `json:"summaries"`
	Dependencies  []parser.Dependency `json:"dependencies"`
}

func WriteJSON(
	path string,
	files []scanner.FileInfo,
	symbols []parser.Symbol,
	summaries map[string]string,
	dependencies []parser.Dependency,
) error {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].FilePath == symbols[j].FilePath {
			return symbols[i].Name < symbols[j].Name
		}
		return symbols[i].FilePath < symbols[j].FilePath
	})

	if dependencies == nil {
		dependencies = []parser.Dependency{}
	}
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].Source != dependencies[j].Source {
			return dependencies[i].Source < dependencies[j].Source
		}
		if dependencies[i].Target != dependencies[j].Target {
			return dependencies[i].Target < dependencies[j].Target
		}
		return dependencies[i].Type < dependencies[j].Type
	})

	pack := ContextPack{
		SchemaVersion: "0.1.0",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Files:         files,
		Symbols:       symbols,
		Summaries:     summaries,
		Dependencies:  dependencies,
	}

	data, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
