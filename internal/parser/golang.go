package parser

import (
	sitter "github.com/smacker/go-tree-sitter"
)

func parseGo(root *sitter.Node, content []byte, path string) []Symbol {
	var symbols []Symbol

	walkGo(root, content, path, &symbols)

	return symbols
}

func walkGo(node *sitter.Node, content []byte, path string, symbols *[]Symbol) {
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
				Language: "go",
			})
		}
	}

	if node.Type() == "method_declaration" {
		nameNode := node.ChildByFieldName("name")

		if nameNode != nil {
			*symbols = append(*symbols, Symbol{
				FilePath: path,
				Name:     nameNode.Content(content),
				Kind:     "method",
				Language: "go",
			})
		}
	}

	if node.Type() == "type_declaration" {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)

			if child.Type() == "type_spec" {
				nameNode := child.ChildByFieldName("name")

				if nameNode != nil {
					*symbols = append(*symbols, Symbol{
						FilePath: path,
						Name:     nameNode.Content(content),
						Kind:     "type",
						Language: "go",
					})
				}
			}
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		walkGo(node.Child(i), content, path, symbols)
	}
}
