package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FileLine represents a specific line in a file
type FileLine struct {
	Path    string
	Line    int
	Content string
}

// ScanResult represents a rule violation
type ScanResult struct {
	Rule       string
	Severity   string // "FAIL" or "WARN"
	File       string
	Line       int
	Message    string
	Suggestion string
}

// GlobFiles returns all files matching the given glob patterns
func GlobFiles(patterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			if !seen[match] {
				seen[match] = true
				files = append(files, match)
			}
		}
	}
	return files, nil
}

// ScanFileForPatterns searches a file for regex patterns and returns matches
func ScanFileForPatterns(path string, patterns []*regexp.Regexp) ([]FileLine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []FileLine
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				matches = append(matches, FileLine{
					Path:    path,
					Line:    lineNum,
					Content: strings.TrimSpace(line),
				})
				break // Only report once per line
			}
		}
	}

	return matches, scanner.Err()
}

// ScanFileForStrings searches a file for string matches
func ScanFileForStrings(path string, searchStrings []string) ([]FileLine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []FileLine
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		for _, s := range searchStrings {
			if strings.Contains(line, s) {
				matches = append(matches, FileLine{
					Path:    path,
					Line:    lineNum,
					Content: strings.TrimSpace(line),
				})
				break
			}
		}
	}

	return matches, scanner.Err()
}

// ExtractImports extracts import statements from a Go file
func ExtractImports(path string) ([]FileLine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var imports []FileLine
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inImportBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Start of import block
		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}

		// End of import block
		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			continue
		}

		// Single import
		if strings.HasPrefix(trimmed, "import ") && !strings.Contains(trimmed, "(") {
			importPath := extractImportPath(trimmed)
			if importPath != "" {
				imports = append(imports, FileLine{
					Path:    path,
					Line:    lineNum,
					Content: importPath,
				})
			}
			continue
		}

		// Inside import block
		if inImportBlock && trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			importPath := extractImportPath(trimmed)
			if importPath != "" {
				imports = append(imports, FileLine{
					Path:    path,
					Line:    lineNum,
					Content: importPath,
				})
			}
		}
	}

	return imports, scanner.Err()
}

// extractImportPath extracts the import path from an import line
func extractImportPath(line string) string {
	// Remove "import" prefix if present
	line = strings.TrimPrefix(line, "import ")

	// Find quoted string
	start := strings.Index(line, `"`)
	if start == -1 {
		return ""
	}
	end := strings.Index(line[start+1:], `"`)
	if end == -1 {
		return ""
	}
	return line[start+1 : start+1+end]
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileContains checks if a file contains a string
func FileContains(path string, search string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), search), nil
}

// ExtractHTMXURLs extracts HTMX endpoint URLs from template files
func ExtractHTMXURLs(path string) ([]FileLine, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []FileLine
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Pattern to match hx-* attributes with URLs
	htmxPattern := regexp.MustCompile(`hx-(get|post|put|delete|patch)="([^"]+)"`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matches := htmxPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				urls = append(urls, FileLine{
					Path:    path,
					Line:    lineNum,
					Content: match[2], // The URL
				})
			}
		}
	}

	return urls, scanner.Err()
}
