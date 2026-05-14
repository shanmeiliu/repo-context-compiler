package parser

import (
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
