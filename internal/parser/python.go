package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func parsePython(root *sitter.Node, content []byte, path string) []Symbol {
	var symbols []Symbol

	walkPython(root, content, path, &symbols)

	return symbols
}

func extractPythonDependencies(root *sitter.Node, content []byte, path string) []Dependency {
	var deps []Dependency

	walkPythonDependencies(root, content, path, &deps)

	return deps
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

func walkPythonDependencies(
	node *sitter.Node,
	content []byte,
	path string,
	deps *[]Dependency,
) {
	if node == nil {
		return
	}

	switch node.Type() {
	case "import_statement":
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			target := pythonImportName(child, content)
			appendDependency(deps, path, target, "python_import")
		}
	case "import_from_statement":
		module := node.ChildByFieldName("module_name")
		if module != nil {
			moduleName := module.Content(content)
			appendDependency(deps, path, moduleName, "python_import")

			for i := 0; i < int(node.NamedChildCount()); i++ {
				child := node.NamedChild(i)
				if child == nil ||
					(child.StartByte() == module.StartByte() && child.EndByte() == module.EndByte()) {
					continue
				}

				importedName := pythonImportName(child, content)
				if importedName == "" {
					continue
				}

				target := strings.TrimSuffix(moduleName, ".") + "." + importedName
				appendDependency(deps, path, target, "python_import")
			}
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		walkPythonDependencies(
			node.Child(i),
			content,
			path,
			deps,
		)
	}
}

func pythonImportName(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}

	if node.Type() == "aliased_import" {
		name := node.ChildByFieldName("name")
		if name != nil {
			return name.Content(content)
		}
	}

	return node.Content(content)
}

func appendDependency(deps *[]Dependency, source, target, dependencyType string) {
	if target == "" {
		return
	}

	*deps = append(*deps, Dependency{
		Source: source,
		Target: target,
		Type:   dependencyType,
	})
}
