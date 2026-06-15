package parser

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shanmeiliu/repo-context-compiler/internal/scanner"
)

func ResolveDependencies(
	repoPath string,
	files []scanner.FileInfo,
	deps []Dependency,
) []Dependency {
	fileSet := make(map[string]scanner.FileInfo, len(files))
	for _, file := range files {
		fileSet[file.Path] = file
	}

	moduleName := readGoModuleName(repoPath)
	var resolved []Dependency
	seen := map[string]bool{}

	for _, dep := range deps {
		targets := resolveDependencyTargets(dep, moduleName, fileSet)
		if len(targets) == 0 {
			targets = []string{dep.Target}
		}

		for _, target := range targets {
			key := dep.Source + "\x00" + target + "\x00" + dep.Type
			if seen[key] {
				continue
			}

			seen[key] = true
			resolved = append(resolved, Dependency{
				Source: dep.Source,
				Target: target,
				Type:   dep.Type,
			})
		}
	}

	sort.Slice(resolved, func(i, j int) bool {
		if resolved[i].Source != resolved[j].Source {
			return resolved[i].Source < resolved[j].Source
		}
		if resolved[i].Target != resolved[j].Target {
			return resolved[i].Target < resolved[j].Target
		}
		return resolved[i].Type < resolved[j].Type
	})

	return resolved
}

func resolveDependencyTargets(
	dep Dependency,
	goModule string,
	files map[string]scanner.FileInfo,
) []string {
	switch dep.Type {
	case "go_import":
		return resolveGoImport(dep.Target, goModule, files)
	case "javascript_import":
		return resolveJavaScriptImport(dep.Source, dep.Target, files)
	case "python_import":
		return resolvePythonImport(dep.Source, dep.Target, files)
	default:
		return nil
	}
}

func resolveGoImport(
	target string,
	moduleName string,
	files map[string]scanner.FileInfo,
) []string {
	if moduleName == "" {
		return nil
	}

	var directory string
	switch {
	case target == moduleName:
		directory = "."
	case strings.HasPrefix(target, moduleName+"/"):
		directory = strings.TrimPrefix(target, moduleName+"/")
	default:
		return nil
	}

	var matches []string
	for filePath, file := range files {
		if file.Language == "go" && path.Dir(filePath) == directory {
			matches = append(matches, filePath)
		}
	}

	sort.Strings(matches)
	return matches
}

func resolveJavaScriptImport(
	source string,
	target string,
	files map[string]scanner.FileInfo,
) []string {
	if !strings.HasPrefix(target, ".") {
		return nil
	}

	base := path.Clean(path.Join(path.Dir(source), target))
	candidates := []string{base}

	if path.Ext(base) == "" {
		for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".json"} {
			candidates = append(candidates, base+ext)
		}
		for _, ext := range []string{".ts", ".tsx", ".js", ".jsx"} {
			candidates = append(candidates, path.Join(base, "index"+ext))
		}
	}

	return existingCandidates(candidates, files)
}

func resolvePythonImport(
	source string,
	target string,
	files map[string]scanner.FileInfo,
) []string {
	dots := 0
	for dots < len(target) && target[dots] == '.' {
		dots++
	}

	module := strings.TrimPrefix(target, strings.Repeat(".", dots))
	modulePath := strings.ReplaceAll(module, ".", "/")

	var base string
	if dots > 0 {
		base = path.Dir(source)
		for i := 1; i < dots; i++ {
			base = path.Dir(base)
		}
		base = path.Join(base, modulePath)
	} else {
		base = modulePath
	}

	candidates := []string{base + ".py", path.Join(base, "__init__.py")}
	if module == "" {
		candidates = []string{path.Join(base, "__init__.py")}
	}

	if matches := existingCandidates(candidates, files); len(matches) > 0 {
		return matches
	}

	var namespacePackageMatches []string
	for filePath, file := range files {
		if file.Language == "python" && path.Dir(filePath) == base && filePath != source {
			namespacePackageMatches = append(namespacePackageMatches, filePath)
		}
	}
	if len(namespacePackageMatches) > 0 {
		sort.Strings(namespacePackageMatches)
		return namespacePackageMatches
	}

	if dots > 0 {
		return nil
	}

	// Python projects commonly keep import roots below directories such as src/.
	var suffixMatches []string
	for _, candidate := range candidates {
		suffix := "/" + candidate
		for filePath := range files {
			if filePath == candidate || strings.HasSuffix(filePath, suffix) {
				suffixMatches = append(suffixMatches, filePath)
			}
		}
	}

	sort.Strings(suffixMatches)
	return dedupeStrings(suffixMatches)
}

func existingCandidates(
	candidates []string,
	files map[string]scanner.FileInfo,
) []string {
	var matches []string
	for _, candidate := range candidates {
		candidate = path.Clean(candidate)
		if _, ok := files[candidate]; ok {
			matches = append(matches, candidate)
		}
	}
	return dedupeStrings(matches)
}

func dedupeStrings(items []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		result = append(result, item)
	}
	return result
}

func readGoModuleName(repoPath string) string {
	data, err := os.ReadFile(filepath.Join(repoPath, "go.mod"))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1]
		}
	}

	return ""
}
