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
	"github.com/shanmeiliu/repo-context-compiler/internal/recommender"
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
	var askLimit int

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
		"",
		"OpenAI-compatible LLM base URL (env: REPOCTX_LLM_BASE_URL)",
	)

	rootCmd.PersistentFlags().StringVar(
		&llmAPIKey,
		"llm-api-key",
		"",
		"LLM API key (env: REPOCTX_LLM_API_KEY)",
	)

	rootCmd.PersistentFlags().StringVar(
		&llmModel,
		"llm-model",
		"",
		"LLM model (env: REPOCTX_LLM_MODEL)",
	)

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if !rootCmd.PersistentFlags().Changed("llm-base-url") {
			llmBaseURL = envOrDefault("REPOCTX_LLM_BASE_URL", "http://localhost:11434/v1")
		}
		if !rootCmd.PersistentFlags().Changed("llm-api-key") {
			llmAPIKey = envOrDefault("REPOCTX_LLM_API_KEY", "")
		}
		if !rootCmd.PersistentFlags().Changed("llm-model") {
			llmModel = envOrDefault("REPOCTX_LLM_MODEL", "qwen2.5-coder:7b")
		}
	}

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
		Short: "Scan repository files, symbols, and dependencies into local database",
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

			symbols := parseSymbols(repoPath, files)
			deps := parseDependencies(repoPath, files)
			deps = parser.ResolveDependencies(repoPath, files, deps)

			if err := db.ReplaceRepositoryState(database, files, symbols, deps); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Scanned %d files\n", len(files))
			fmt.Printf("Extracted %d symbols\n", len(symbols))
			fmt.Printf("Extracted %d dependencies\n", len(deps))
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

			symbols := parseSymbols(repoPath, files)
			deps := parseDependencies(repoPath, files)
			deps = parser.ResolveDependencies(repoPath, files, deps)

			if err := db.ReplaceRepositoryState(database, files, symbols, deps); err != nil {
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

			repoName := filepath.Base(repoPath)
			if absoluteRepoPath, err := filepath.Abs(repoPath); err == nil {
				repoName = filepath.Base(absoluteRepoPath)
			}

			if err := pack.WriteMarkdown(
				outDir+"/ai-context.md",
				repoName,
				files,
				symbols,
				summaries,
				deps,
			); err != nil {
				log.Fatal(err)
			}

			if err := pack.WriteJSON(outDir+"/ai-context.json", files, symbols, summaries, deps); err != nil {
				log.Fatal(err)
			}

			fmt.Println("Generated:", outDir+"/ai-context.md")
			fmt.Println("Generated:", outDir+"/ai-context.json")
		},
	}

	var askCmd = &cobra.Command{
		Use:   "ask [task]",
		Short: "Recommend relevant files for a task using the local context database",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			task := args[0]

			database, err := db.Open(dbPath)
			if err != nil {
				log.Fatal(err)
			}
			defer database.Close()

			if err := db.Migrate(database); err != nil {
				log.Fatal(err)
			}

			files, err := db.GetFiles(database)
			if err != nil {
				log.Fatal(err)
			}

			symbols, err := db.GetSymbols(database)
			if err != nil {
				log.Fatal(err)
			}

			deps, err := db.GetDependencies(database)
			if err != nil {
				log.Fatal(err)
			}

			recs := recommender.RecommendFiles(task, files, symbols, deps, askLimit)

			if len(recs) == 0 {
				fmt.Println("No likely files found. Try running `repoctx pack . --summarize` first for better recommendations.")
				return
			}

			fmt.Println("Likely relevant files:")
			fmt.Println()

			for _, rec := range recs {
				fmt.Printf("- %s (score: %d)\n", rec.Path, rec.Score)

				for _, reason := range rec.Reasons {
					fmt.Printf("  - %s\n", reason)
				}
			}
		},
	}

	askCmd.Flags().IntVar(&askLimit, "limit", 10, "Maximum number of files to recommend")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(askCmd)

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

func parseDependencies(
	repoPath string,
	files []scanner.FileInfo,
) []parser.Dependency {
	var deps []parser.Dependency

	for _, file := range files {
		fullPath := filepath.Join(repoPath, file.Path)

		parsed, err := parser.ParseDependencies(fullPath)
		if err != nil {
			continue
		}

		for i := range parsed {
			parsed[i].Source = file.Path
		}

		deps = append(deps, parsed...)
	}

	return deps
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

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
