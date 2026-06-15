package parser

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	golang "github.com/smacker/go-tree-sitter/golang"
	javascript "github.com/smacker/go-tree-sitter/javascript"
	python "github.com/smacker/go-tree-sitter/python"
	tsx "github.com/smacker/go-tree-sitter/typescript/tsx"
	typescript "github.com/smacker/go-tree-sitter/typescript/typescript"
)

type Symbol struct {
	FilePath     string   `json:"file_path"`
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	Language     string   `json:"language"`
	Dependencies []string `json:"dependencies"`
}

func ParseDependencies(path string) (deps []Dependency, err error) {
	defer func() {
		if recover() != nil {
			deps = nil
			err = nil
		}
	}()

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".py":
		return parseTreeDependencies(path, content, python.GetLanguage(), extractPythonDependencies)
	case ".js", ".jsx":
		return parseTreeDependencies(path, content, javascript.GetLanguage(), extractJavaScriptDependencies)
	case ".ts":
		return parseTreeDependencies(path, content, typescript.GetLanguage(), extractJavaScriptDependencies)
	case ".tsx":
		return parseTreeDependencies(path, content, tsx.GetLanguage(), extractJavaScriptDependencies)
	case ".go":
		return extractGoDependencies(path, content)
	default:
		return nil, nil
	}
}

func parseTreeDependencies(
	path string,
	content []byte,
	lang *sitter.Language,
	extract func(*sitter.Node, []byte, string) []Dependency,
) ([]Dependency, error) {
	if lang == nil {
		return nil, nil
	}

	p := sitter.NewParser()
	if p == nil {
		return nil, nil
	}

	p.SetLanguage(lang)

	tree, err := p.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}

	if tree == nil {
		return nil, nil
	}

	root := tree.RootNode()
	if root == nil {
		return nil, nil
	}

	return extract(root, content, path), nil
}

func ParseFile(path string) (symbols []Symbol, err error) {
	// Defensive guard: tree-sitter is CGO-backed and can panic on bad parser state.
	defer func() {
		if recover() != nil {
			symbols = nil
			err = nil
		}
	}()

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))

	var lang *sitter.Language

	switch ext {
	case ".py":
		lang = python.GetLanguage()
	case ".js", ".jsx":
		lang = javascript.GetLanguage()
	case ".ts":
		lang = typescript.GetLanguage()
	case ".tsx":
		lang = tsx.GetLanguage()
	case ".go":
		lang = golang.GetLanguage()
	default:
		return nil, nil
	}

	if lang == nil {
		return nil, nil
	}

	p := sitter.NewParser()
	if p == nil {
		return nil, nil
	}

	p.SetLanguage(lang)

	tree, err := p.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}

	if tree == nil {
		return nil, nil
	}

	root := tree.RootNode()
	if root == nil {
		return nil, nil
	}

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
