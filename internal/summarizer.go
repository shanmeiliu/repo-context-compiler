package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/shanmeiliu/repo-context-compiler/internal/llm"
	"github.com/shanmeiliu/repo-context-compiler/internal/parser"
	"github.com/shanmeiliu/repo-context-compiler/internal/scanner"
)

type FileSummary struct {
	Path    string `json:"path"`
	Summary string `json:"summary"`
}

func SummarizeFiles(
	repoPath string,
	files []scanner.FileInfo,
	symbols []parser.Symbol,
	client *llm.Client,
	maxBytes int64,
) ([]FileSummary, error) {
	symbolsByFile := map[string][]parser.Symbol{}

	for _, symbol := range symbols {
		symbolsByFile[symbol.FilePath] = append(symbolsByFile[symbol.FilePath], symbol)
	}

	var summaries []FileSummary

	for _, file := range files {
		if !shouldSummarize(file) {
			continue
		}

		if file.SizeBytes > maxBytes {
			summaries = append(summaries, FileSummary{
				Path:    file.Path,
				Summary: "Skipped: file is too large for local summarization.",
			})
			continue
		}

		fullPath := repoPath + "/" + file.Path

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		prompt := buildFileSummaryPrompt(file, symbolsByFile[file.Path], string(content))

		summary, err := client.Generate(prompt)
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, FileSummary{
			Path:    file.Path,
			Summary: strings.TrimSpace(summary),
		})
	}

	return summaries, nil
}

func shouldSummarize(file scanner.FileInfo) bool {
	switch file.Language {
	case "go", "python", "typescript", "typescriptreact", "javascript", "javascriptreact", "sql":
		return true
	default:
		return false
	}
}

func buildFileSummaryPrompt(file scanner.FileInfo, symbols []parser.Symbol, content string) string {
	var symbolLines []string

	for _, symbol := range symbols {
		symbolLines = append(symbolLines, fmt.Sprintf("- %s: %s", symbol.Kind, symbol.Name))
	}

	return fmt.Sprintf(`
You are analyzing a source code file for an AI coding assistant.

Return a concise but useful technical summary.

Focus on:
- purpose of the file
- major functions/classes
- dependencies
- side effects
- risk level if edited
- what future LLMs should know before modifying this file

File path:
%s

Language:
%s

Detected symbols:
%s

Source code:
%s

Return format:

Purpose:
Dependencies:
Side Effects:
Risk:
Notes:
`, file.Path, file.Language, strings.Join(symbolLines, "\n"), content)
}
