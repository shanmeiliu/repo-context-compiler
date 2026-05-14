package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Path      string `json:"path"`
	Language  string `json:"language"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".venv":        true,
	"venv":         true,
	"__pycache__":  true,
	".repoctx":     true,
}

func Scan(root string) ([]FileInfo, error) {
	var files []FileInfo

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if ignoredDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldIgnoreFile(path) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		hash, err := hashFile(path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}

		files = append(files, FileInfo{
			Path:      filepath.ToSlash(rel),
			Language:  detectLanguage(path),
			SizeBytes: info.Size(),
			SHA256:    hash,
		})

		return nil
	})

	return files, err
}

func shouldIgnoreFile(path string) bool {
	name := filepath.Base(path)

	if strings.HasPrefix(name, ".DS_Store") {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))

	ignoredExts := map[string]bool{
		".png":    true,
		".jpg":    true,
		".jpeg":   true,
		".gif":    true,
		".ico":    true,
		".pdf":    true,
		".zip":    true,
		".gz":     true,
		".db":     true,
		".sqlite": true,
	}

	return ignoredExts[ext]
}

func detectLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescriptreact"
	case ".js":
		return "javascript"
	case ".jsx":
		return "javascriptreact"
	case ".json":
		return "json"
	case ".md":
		return "markdown"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".sql":
		return "sql"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "unknown"
	}
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()

	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
