package parser

import (
	"os"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
	golang "github.com/smacker/go-tree-sitter/golang"
	javascript "github.com/smacker/go-tree-sitter/javascript"
	python "github.com/smacker/go-tree-sitter/python"
	typescript "github.com/smacker/go-tree-sitter/typescript/typescript"
)

type Symbol struct {
	FilePath     string   `json:"file_path"`
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	Language     string   `json:"language"`
	Dependencies []string `json:"dependencies"`
}

func ParseFile(path string) ([]Symbol, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(path)

	var lang *sitter.Language

	switch ext {
	case ".py":
		lang = python.GetLanguage()
	case ".js", ".jsx":
		lang = javascript.GetLanguage()
	case ".ts", ".tsx":
		lang = typescript.GetLanguage()
	case ".go":
		lang = golang.GetLanguage()
	default:
		return nil, nil
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(nil, nil, content)
	if err != nil {
		return nil, err
	}

	root := tree.RootNode()

	switch ext {
	case ".py":
		return parsePython(root, content, path), nil
	case ".js", ".jsx", ".ts", ".tsx":
		return parseJavaScript(root, content, path), nil
	case ".go":
		return parseGo(root, content, path), nil
	default:
		return nil, nil
	}
}
