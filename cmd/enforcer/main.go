package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Running NCOE Enforcer...")
	fmt.Println()

	report := &Report{}

	// R1: Check HTMX fragment URLs use "/_" convention
	checkHTMXFragmentURLs(report)

	// R2: Check handler imports
	checkHandlerImports(report)

	// R3: Check repository imports
	checkRepoImports(report)

	// R5: Check request_id middleware
	checkRequestIDMiddleware(report)

	// Print report
	report.Print()

	// Exit with appropriate code
	if report.HasFailures() {
		os.Exit(1)
	}
}

// R1: All HTMX fragment URLs must include "/_"
func checkHTMXFragmentURLs(report *Report) {
	files, err := GlobFiles(TemplatePaths)
	if err != nil {
		fmt.Printf("Warning: could not glob template files: %v\n", err)
		return
	}

	for _, file := range files {
		urls, err := ExtractHTMXURLs(file)
		if err != nil {
			fmt.Printf("Warning: could not scan %s: %v\n", file, err)
			continue
		}

		for _, url := range urls {
			// Skip if URL is in allowlist
			if isAllowedNonFragment(url.Content) {
				continue
			}

			// Check if it's a fragment URL (contains staff/ and a fragment suffix)
			if isFragmentURL(url.Content) && !strings.Contains(url.Content, "/_") {
				report.AddFailure(
					"R1",
					url.Path,
					url.Line,
					fmt.Sprintf("HTMX fragment URL missing '/_': %s", url.Content),
					fmt.Sprintf("Change to include '/_' before fragment suffix (e.g., /_panel, /_table)"),
				)
			}
		}
	}
}

// isFragmentURL checks if a URL looks like a fragment endpoint
func isFragmentURL(url string) bool {
	// Check if URL contains any fragment suffix pattern
	for _, suffix := range FragmentRouteSuffixes {
		if strings.Contains(url, suffix) {
			return true
		}
	}
	return false
}

// isAllowedNonFragment checks if URL is in the allowlist
func isAllowedNonFragment(url string) bool {
	for _, allowed := range AllowedNonFragmentURLs {
		if strings.Contains(url, allowed) {
			return true
		}
	}
	return false
}

// R2: Handlers must not import forbidden packages
func checkHandlerImports(report *Report) {
	files, err := GlobFiles(HandlerPaths)
	if err != nil {
		fmt.Printf("Warning: could not glob handler files: %v\n", err)
		return
	}

	for _, file := range files {
		imports, err := ExtractImports(file)
		if err != nil {
			fmt.Printf("Warning: could not extract imports from %s: %v\n", file, err)
			continue
		}

		for _, imp := range imports {
			for _, forbidden := range HandlerForbiddenImports {
				if strings.Contains(imp.Content, forbidden) {
					report.AddFailure(
						"R2",
						imp.Path,
						imp.Line,
						fmt.Sprintf("Handler imports forbidden package: %s", imp.Content),
						"Handlers should only do HTTP I/O. Move business logic to services.",
					)
				}
			}
		}
	}
}

// R3: Repositories must not import forbidden packages
func checkRepoImports(report *Report) {
	files, err := GlobFiles(RepoPaths)
	if err != nil {
		fmt.Printf("Warning: could not glob repository files: %v\n", err)
		return
	}

	for _, file := range files {
		imports, err := ExtractImports(file)
		if err != nil {
			fmt.Printf("Warning: could not extract imports from %s: %v\n", file, err)
			continue
		}

		for _, imp := range imports {
			for _, forbidden := range RepoForbiddenImports {
				if strings.Contains(imp.Content, forbidden) {
					report.AddFailure(
						"R3",
						imp.Path,
						imp.Line,
						fmt.Sprintf("Repository imports forbidden package: %s", imp.Content),
						"Repositories should only do SQL. Move HTTP logic to handlers.",
					)
				}
			}
		}
	}
}

// R5: request_id middleware must exist and be wired
func checkRequestIDMiddleware(report *Report) {
	// Check if middleware file exists
	if !FileExists(RequestIDRequirements.MiddlewareFile) {
		report.AddFailure(
			"R5",
			RequestIDRequirements.MiddlewareFile,
			0,
			"request_id middleware file does not exist",
			fmt.Sprintf("Create %s with RequestID middleware function", RequestIDRequirements.MiddlewareFile),
		)
		return
	}

	// Check if RequestID function exists in the file
	hasFunc, err := FileContains(RequestIDRequirements.MiddlewareFile, "func "+RequestIDRequirements.FunctionName)
	if err != nil {
		fmt.Printf("Warning: could not read %s: %v\n", RequestIDRequirements.MiddlewareFile, err)
		return
	}
	if !hasFunc {
		report.AddFailure(
			"R5",
			RequestIDRequirements.MiddlewareFile,
			0,
			fmt.Sprintf("RequestID function not found in middleware"),
			"Add func RequestID(next http.Handler) http.Handler",
		)
	}

	// Check if main.go wires the middleware
	if !FileExists(MainGoPath) {
		report.AddFailure(
			"R5",
			MainGoPath,
			0,
			"main.go not found",
			"Ensure main.go exists and wires RequestID middleware",
		)
		return
	}

	mainHasRequestID, err := FileContains(MainGoPath, "middleware.RequestID")
	if err != nil {
		fmt.Printf("Warning: could not read %s: %v\n", MainGoPath, err)
		return
	}
	if !mainHasRequestID {
		report.AddFailure(
			"R5",
			MainGoPath,
			0,
			"RequestID middleware not wired in main.go",
			"Add middleware.RequestID to the middleware chain",
		)
	}

	// Check if X-Request-Id header is set
	hasHeader, err := FileContains(RequestIDRequirements.MiddlewareFile, RequestIDRequirements.HeaderName)
	if err == nil && !hasHeader {
		report.AddFailure(
			"R5",
			RequestIDRequirements.MiddlewareFile,
			0,
			fmt.Sprintf("%s header not set in middleware", RequestIDRequirements.HeaderName),
			fmt.Sprintf("Add w.Header().Set(\"%s\", requestID)", RequestIDRequirements.HeaderName),
		)
	}
}
