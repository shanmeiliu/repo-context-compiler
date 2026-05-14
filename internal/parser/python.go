package parser

import (
	sitter "github.com/smacker/go-tree-sitter"
)

func parsePython(root *sitter.Node, content []byte, path string) []Symbol {
	var symbols []Symbol

	walkPython(root, content, path, &symbols)

	return symbols
}

func walkPython(node *sitter.Node, content []byte, path string, symbols *[]Symbol) {
	if node == nil {
		return
	}

	if node.Type() == "function_definition" {
		nameNode := node.ChildByFieldName("name")

		if nameNode != nil {
			*symbols = append(*symbols, Symbol{
				FilePath: path,
				Name:     nameNode.Content(content),
				Kind:     "function",
				Language: "python",
			})
		}
	}

	if node.Type() == "class_definition" {
		nameNode := node.ChildByFieldName("name")

		if nameNode != nil {
			*symbols = append(*symbols, Symbol{
				FilePath: path,
				Name:     nameNode.Content(content),
				Kind:     "class",
				Language: "python",
			})
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		walkPython(node.Child(i), content, path, symbols)
	}
}
