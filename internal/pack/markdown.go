package pack

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shanmeiliu/repo-context-compiler/internal/parser"
	"github.com/shanmeiliu/repo-context-compiler/internal/scanner"
)

func WriteMarkdown(
	path string,
	repoName string,
	files []scanner.FileInfo,
	symbols []parser.Symbol,
	summaries map[string]string,
	dependencies []parser.Dependency,
) error {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	var b strings.Builder

	b.WriteString("# AI Context Pack\n\n")
	b.WriteString("## Repo Summary\n\n")
	b.WriteString("This context pack was generated locally by `repo-context-compiler`.\n\n")

	b.WriteString("## Repository Tree\n\n")
	b.WriteString("```text\n")
	writeRepositoryTree(&b, repoName, files)
	b.WriteString("```\n\n")

	b.WriteString("## Files\n\n")

	for _, file := range files {
		b.WriteString(fmt.Sprintf("### `%s`\n\n", file.Path))
		b.WriteString(fmt.Sprintf("- Language: `%s`\n", file.Language))
		b.WriteString(fmt.Sprintf("- Size: `%d bytes`\n", file.SizeBytes))
		b.WriteString(fmt.Sprintf("- SHA256: `%s`\n\n", file.SHA256))

		summary := summaries[file.Path]
		if summary == "" {
			summary = "_Not generated yet._"
		}

		b.WriteString("Summary:\n\n")
		b.WriteString(summary)
		b.WriteString("\n\n")
	}

	b.WriteString("## Symbols\n\n")

	symbolsByFile := make(map[string][]parser.Symbol)

	for _, symbol := range symbols {
		symbolsByFile[symbol.FilePath] = append(symbolsByFile[symbol.FilePath], symbol)
	}

	for _, file := range files {
		fileSymbols := symbolsByFile[file.Path]
		if len(fileSymbols) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("### `%s`\n\n", file.Path))

		for _, symbol := range fileSymbols {
			b.WriteString(fmt.Sprintf("- `%s`: `%s`\n", symbol.Kind, symbol.Name))
		}

		b.WriteString("\n")
	}

	b.WriteString("## Dependencies\n\n")

	if len(dependencies) == 0 {
		b.WriteString("_No dependencies detected._\n")
	} else {
		sort.Slice(dependencies, func(i, j int) bool {
			if dependencies[i].Source != dependencies[j].Source {
				return dependencies[i].Source < dependencies[j].Source
			}
			if dependencies[i].Target != dependencies[j].Target {
				return dependencies[i].Target < dependencies[j].Target
			}
			return dependencies[i].Type < dependencies[j].Type
		})

		for _, dependency := range dependencies {
			b.WriteString(fmt.Sprintf(
				"- `%s` -> `%s` (`%s`)\n",
				dependency.Source,
				dependency.Target,
				dependency.Type,
			))
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

type treeNode struct {
	name     string
	isFile   bool
	children map[string]*treeNode
}

func writeRepositoryTree(
	b *strings.Builder,
	repoName string,
	files []scanner.FileInfo,
) {
	root := &treeNode{children: map[string]*treeNode{}}

	for _, file := range files {
		parts := strings.Split(file.Path, "/")
		current := root

		for index, part := range parts {
			child, exists := current.children[part]
			if !exists {
				child = &treeNode{
					name:     part,
					isFile:   index == len(parts)-1,
					children: map[string]*treeNode{},
				}
				current.children[part] = child
			}
			current = child
		}
	}

	if repoName == "" {
		repoName = "."
	}
	b.WriteString(repoName)
	b.WriteString("/\n")
	writeTreeChildren(b, root, "")
}

func writeTreeChildren(b *strings.Builder, node *treeNode, prefix string) {
	children := make([]*treeNode, 0, len(node.children))
	for _, child := range node.children {
		children = append(children, child)
	}

	sort.Slice(children, func(i, j int) bool {
		if children[i].isFile != children[j].isFile {
			return !children[i].isFile
		}
		return children[i].name < children[j].name
	})

	for index, child := range children {
		last := index == len(children)-1
		branch := "├── "
		childPrefix := prefix + "│   "
		if last {
			branch = "└── "
			childPrefix = prefix + "    "
		}

		b.WriteString(prefix)
		b.WriteString(branch)
		b.WriteString(child.name)
		if !child.isFile {
			b.WriteString("/")
		}
		b.WriteString("\n")

		if !child.isFile {
			writeTreeChildren(b, child, childPrefix)
		}
	}
}
