package pack

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shanmeiliu/repo-context-compiler/internal/scanner"
)

func WriteMarkdown(path string, files []scanner.FileInfo) error {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	var b strings.Builder

	b.WriteString("# AI Context Pack\n\n")
	b.WriteString("## Repo Summary\n\n")
	b.WriteString("This context pack was generated locally by `repo-context-compiler`.\n\n")

	b.WriteString("## Repository Tree\n\n")
	b.WriteString("```text\n")
	for _, file := range files {
		b.WriteString(file.Path)
		b.WriteString("\n")
	}
	b.WriteString("```\n\n")

	b.WriteString("## Files\n\n")

	for _, file := range files {
		b.WriteString(fmt.Sprintf("### `%s`\n\n", file.Path))
		b.WriteString(fmt.Sprintf("- Language: `%s`\n", file.Language))
		b.WriteString(fmt.Sprintf("- Size: `%d bytes`\n", file.SizeBytes))
		b.WriteString(fmt.Sprintf("- SHA256: `%s`\n\n", file.SHA256))
		b.WriteString("Summary: _Not generated yet._\n\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}
