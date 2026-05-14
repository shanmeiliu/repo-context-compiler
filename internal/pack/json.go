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
	SchemaVersion string             `json:"schema_version"`
	GeneratedAt   string             `json:"generated_at"`
	Files         []scanner.FileInfo `json:"files"`
	Symbols       []Symbol           `json:"symbols"`
}
type Symbol struct {
	FilePath string `json:"file_path"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Language string `json:"language"`
}

func WriteJSON(path string, files []scanner.FileInfo, symbols []parser.Symbol) error {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	pack := ContextPack{
		SchemaVersion: "0.1.0",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Files:         files,
		Symbols:       symbols,
	}

	data, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
