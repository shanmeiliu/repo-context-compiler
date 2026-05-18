package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	repoctx "github.com/shanmeiliu/repo-context-compiler/internal"
	"github.com/shanmeiliu/repo-context-compiler/internal/db"
	"github.com/shanmeiliu/repo-context-compiler/internal/llm"
	"github.com/shanmeiliu/repo-context-compiler/internal/pack"
	"github.com/shanmeiliu/repo-context-compiler/internal/parser"
	"github.com/shanmeiliu/repo-context-compiler/internal/scanner"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	_ = godotenv.Load()

	var dbPath string
	var outDir string
	var summarize bool
	var llmBaseURL string
	var llmAPIKey string
	var llmModel string

	var rootCmd = &cobra.Command{
		Use:   "repoctx",
		Short: "Compile a repository into an AI-friendly context pack",
	}

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", ".repoctx/repoctx.db", "SQLite database path")
	rootCmd.PersistentFlags().StringVar(&outDir, "out", ".repoctx", "Output directory")
	rootCmd.PersistentFlags().BoolVar(&summarize, "summarize", false, "Generate local or remote LLM file summaries")

	rootCmd.PersistentFlags().StringVar(
		&llmBaseURL,
		"llm-base-url",
		envOrDefault("REPOCTX_LLM_BASE_URL", "http://localhost:11434/v1"),
		"OpenAI-compatible LLM base URL",
	)

	rootCmd.PersistentFlags().StringVar(
		&llmAPIKey,
		"llm-api-key",
		envOrDefault("REPOCTX_LLM_API_KEY", ""),
		"LLM API key",
	)

	rootCmd.PersistentFlags().StringVar(
		&llmModel,
		"llm-model",
		envOrDefault("REPOCTX_LLM_MODEL", "qwen2.5-coder:7b"),
		"LLM model",
	)

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize local repo context database",
		Run: func(cmd *cobra.Command, args []string) {
			database, err := db.Open(dbPath)
			if err != nil {
				log.Fatal(err)
			}
			defer database.Close()

			if err := db.Migrate(database); err != nil {
				log.Fatal(err)
			}

			fmt.Println("Initialized:", dbPath)
		},
	}

	var scanCmd = &cobra.Command{
		Use:   "scan [repo path]",
		Short: "Scan repository files and symbols into local database",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repoPath := args[0]

			database, err := db.Open(dbPath)
			if err != nil {
				log.Fatal(err)
			}
			defer database.Close()

			if err := db.Migrate(database); err != nil {
				log.Fatal(err)
			}

			files, err := scanner.Scan(repoPath)
			if err != nil {
				log.Fatal(err)
			}

			if err := db.UpsertFiles(database, files); err != nil {
				log.Fatal(err)
			}

			symbols := parseSymbols(repoPath, files)

			if err := db.UpsertSymbols(database, symbols); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Scanned %d files\n", len(files))
			fmt.Printf("Extracted %d symbols\n", len(symbols))
		},
	}

	var packCmd = &cobra.Command{
		Use:   "pack [repo path]",
		Short: "Generate AI-friendly context files",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			repoPath := args[0]

			database, err := db.Open(dbPath)
			if err != nil {
				log.Fatal(err)
			}
			defer database.Close()

			if err := db.Migrate(database); err != nil {
				log.Fatal(err)
			}

			previousState, err := db.GetFileSummaryState(database)
			if err != nil {
				log.Fatal(err)
			}

			files, err := scanner.Scan(repoPath)
			if err != nil {
				log.Fatal(err)
			}

			if err := db.UpsertFiles(database, files); err != nil {
				log.Fatal(err)
			}

			symbols := parseSymbols(repoPath, files)

			if err := db.UpsertSymbols(database, symbols); err != nil {
				log.Fatal(err)
			}

			if summarize {
				filesToSummarize := filterFilesNeedingSummaries(files, previousState)

				if len(filesToSummarize) == 0 {
					fmt.Println("No file summaries needed; using cached summaries")
				} else {
					client := llm.NewClient(llmBaseURL, llmAPIKey, llmModel)

					fileSummaries, err := repoctx.SummarizeFiles(
						repoPath,
						filesToSummarize,
						symbols,
						client,
						12000,
					)
					if err != nil {
						log.Fatal(err)
					}

					if err := db.UpsertFileSummaries(database, fileSummaries); err != nil {
						log.Fatal(err)
					}

					fmt.Printf("Generated %d file summaries\n", len(fileSummaries))
					fmt.Printf("Reused %d cached summaries\n", len(files)-len(filesToSummarize))
				}
			}

			summaries, err := db.GetFileSummaries(database)
			if err != nil {
				log.Fatal(err)
			}

			if err := os.MkdirAll(outDir, 0755); err != nil {
				log.Fatal(err)
			}

			if err := pack.WriteMarkdown(outDir+"/ai-context.md", files, symbols, summaries); err != nil {
				log.Fatal(err)
			}

			if err := pack.WriteJSON(outDir+"/ai-context.json", files, symbols, summaries); err != nil {
				log.Fatal(err)
			}

			fmt.Println("Generated:", outDir+"/ai-context.md")
			fmt.Println("Generated:", outDir+"/ai-context.json")
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(packCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func parseSymbols(repoPath string, files []scanner.FileInfo) []parser.Symbol {
	var symbols []parser.Symbol

	for _, file := range files {
		fullPath := filepath.Join(repoPath, file.Path)

		parsed, err := parser.ParseFile(fullPath)
		if err != nil {
			continue
		}

		for i := range parsed {
			parsed[i].FilePath = file.Path
		}

		symbols = append(symbols, parsed...)
	}

	return symbols
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func filterFilesNeedingSummaries(
	files []scanner.FileInfo,
	previousState map[string]db.FileSummaryState,
) []scanner.FileInfo {
	var result []scanner.FileInfo

	for _, file := range files {
		old, exists := previousState[file.Path]

		if !exists {
			result = append(result, file)
			continue
		}

		if old.SHA256 != file.SHA256 {
			result = append(result, file)
			continue
		}

		if !old.HasSummary {
			result = append(result, file)
			continue
		}
	}

	return result
}
