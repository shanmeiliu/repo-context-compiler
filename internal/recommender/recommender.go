package recommender

import (
	"sort"
	"strings"

	"github.com/shanmeiliu/repo-context-compiler/internal/db"
)

type Recommendation struct {
	Path    string
	Score   int
	Reasons []string
}

func RecommendFiles(
	task string,
	files []db.FileRecord,
	symbols []db.SymbolRecord,
	deps []db.DependencyRecord,
	limit int,
) []Recommendation {
	taskTerms := tokenize(task)

	symbolsByFile := map[string][]db.SymbolRecord{}
	knownFiles := map[string]bool{}

	for _, symbol := range symbols {
		symbolsByFile[symbol.FilePath] = append(symbolsByFile[symbol.FilePath], symbol)
	}
	for _, file := range files {
		knownFiles[file.Path] = true
	}

	recs := map[string]*Recommendation{}

	for _, file := range files {
		score := 0
		var reasons []string

		pathLower := strings.ToLower(file.Path)
		summaryLower := strings.ToLower(file.Summary)

		for _, term := range taskTerms {
			if strings.Contains(pathLower, term) {
				score += 5
				reasons = append(reasons, "path matches task term: "+term)
			}

			if summaryLower != "" && strings.Contains(summaryLower, term) {
				score += 4
				reasons = append(reasons, "summary matches task term: "+term)
			}
		}

		for _, symbol := range symbolsByFile[file.Path] {
			symbolLower := strings.ToLower(symbol.Name)

			for _, term := range taskTerms {
				if strings.Contains(symbolLower, term) {
					score += 3
					reasons = append(reasons, "symbol matches task term: "+symbol.Name)
				}
			}
		}

		if score > 0 {
			recs[file.Path] = &Recommendation{
				Path:    file.Path,
				Score:   score,
				Reasons: dedupe(reasons),
			}
		}
	}

	// Boost direct dependency neighbors.
	for _, dep := range deps {
		sourceRec, sourceMatched := recs[dep.Source]
		if sourceMatched {
			targetPath := strings.TrimSpace(dep.Target)

			if knownFiles[targetPath] {
				rec, ok := recs[targetPath]
				if !ok {
					rec = &Recommendation{
						Path: targetPath,
					}
					recs[targetPath] = rec
				}

				rec.Score += 2
				rec.Reasons = append(rec.Reasons, "dependency neighbor of "+sourceRec.Path)
			}
		}
	}

	var result []Recommendation

	for _, rec := range recs {
		rec.Reasons = dedupe(rec.Reasons)
		result = append(result, *rec)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Path < result[j].Path
		}

		return result[i].Score > result[j].Score
	})

	if limit > 0 && len(result) > limit {
		return result[:limit]
	}

	return result
}

func tokenize(input string) []string {
	input = strings.ToLower(input)

	replacer := strings.NewReplacer(
		"_", " ",
		"-", " ",
		".", " ",
		"/", " ",
		"\\", " ",
		":", " ",
		"(", " ",
		")", " ",
		",", " ",
	)

	input = replacer.Replace(input)

	raw := strings.Fields(input)

	stopwords := map[string]bool{
		"a": true, "an": true, "the": true,
		"to": true, "of": true, "in": true,
		"on": true, "for": true, "and": true,
		"or": true, "with": true, "after": true,
		"before": true, "is": true, "are": true,
		"fix": true, "bug": true, "issue": true,
		"add": true, "update": true, "change": true,
	}

	var terms []string

	for _, word := range raw {
		word = strings.TrimSpace(word)

		if len(word) < 3 {
			continue
		}

		if stopwords[word] {
			continue
		}

		terms = append(terms, word)
	}

	return dedupe(terms)
}

func dedupe(items []string) []string {
	seen := map[string]bool{}
	var result []string

	for _, item := range items {
		if item == "" {
			continue
		}

		if seen[item] {
			continue
		}

		seen[item] = true
		result = append(result, item)
	}

	return result
}
