package parser

import (
	"strconv"

	sitter "github.com/smacker/go-tree-sitter"
)

func parseJavaScript(root *sitter.Node, content []byte, path string) []Symbol {
	var symbols []Symbol

	walkJavaScript(root, content, path, &symbols)

	return symbols
}

func walkJavaScript(node *sitter.Node, content []byte, path string, symbols *[]Symbol) {
	if node == nil {
		return
	}

	if node.Type() == "function_declaration" {
		nameNode := node.ChildByFieldName("name")

		if nameNode != nil {
			*symbols = append(*symbols, Symbol{
				FilePath: path,
				Name:     nameNode.Content(content),
				Kind:     "function",
				Language: "javascript",
			})
		}
	}

	if node.Type() == "class_declaration" {
		nameNode := node.ChildByFieldName("name")

		if nameNode != nil {
			*symbols = append(*symbols, Symbol{
				FilePath: path,
				Name:     nameNode.Content(content),
				Kind:     "class",
				Language: "javascript",
			})
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		walkJavaScript(node.Child(i), content, path, symbols)
	}
}

func extractJavaScriptDependencies(root *sitter.Node, content []byte, path string) []Dependency {
	var deps []Dependency
	walkJavaScriptDependencies(root, content, path, &deps)
	return deps
}

func walkJavaScriptDependencies(
	node *sitter.Node,
	content []byte,
	path string,
	deps *[]Dependency,
) {
	if node == nil {
		return
	}

	if node.Type() == "import_statement" || node.Type() == "export_statement" {
		source := node.ChildByFieldName("source")
		if source != nil {
			target, err := strconv.Unquote(source.Content(content))
			if err == nil {
				*deps = append(*deps, Dependency{
					Source: path,
					Target: target,
					Type:   "javascript_import",
				})
			}
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		walkJavaScriptDependencies(node.Child(i), content, path, deps)
	}
}
