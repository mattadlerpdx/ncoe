package main

import (
	"fmt"
	"strings"
)

// Report holds all scan results
type Report struct {
	Failures []ScanResult
	Warnings []ScanResult
}

// AddFailure adds a failure to the report
func (r *Report) AddFailure(rule, file string, line int, message, suggestion string) {
	r.Failures = append(r.Failures, ScanResult{
		Rule:       rule,
		Severity:   "FAIL",
		File:       file,
		Line:       line,
		Message:    message,
		Suggestion: suggestion,
	})
}

// AddWarning adds a warning to the report
func (r *Report) AddWarning(rule, file string, line int, message, suggestion string) {
	r.Warnings = append(r.Warnings, ScanResult{
		Rule:       rule,
		Severity:   "WARN",
		File:       file,
		Line:       line,
		Message:    message,
		Suggestion: suggestion,
	})
}

// HasFailures returns true if there are any failures
func (r *Report) HasFailures() bool {
	return len(r.Failures) > 0
}

// Print outputs the report to stdout
func (r *Report) Print() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("ENFORCER REPORT")
	fmt.Println(strings.Repeat("=", 70))

	if len(r.Failures) > 0 {
		fmt.Printf("\n❌ FAILURES (%d)\n", len(r.Failures))
		fmt.Println(strings.Repeat("-", 50))
		for _, f := range r.Failures {
			fmt.Printf("\n[%s] %s:%d\n", f.Rule, f.File, f.Line)
			fmt.Printf("  Message: %s\n", f.Message)
			if f.Suggestion != "" {
				fmt.Printf("  Fix: %s\n", f.Suggestion)
			}
		}
	}

	if len(r.Warnings) > 0 {
		fmt.Printf("\n⚠️  WARNINGS (%d)\n", len(r.Warnings))
		fmt.Println(strings.Repeat("-", 50))
		for _, w := range r.Warnings {
			fmt.Printf("\n[%s] %s:%d\n", w.Rule, w.File, w.Line)
			fmt.Printf("  Message: %s\n", w.Message)
			if w.Suggestion != "" {
				fmt.Printf("  Suggestion: %s\n", w.Suggestion)
			}
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))

	if !r.HasFailures() {
		fmt.Println("✅ ALL CHECKS PASSED")
		fmt.Println("   ✓ R1: All HTMX fragment URLs use '/_' convention")
		fmt.Println("   ✓ R2: Handlers have no forbidden imports")
		fmt.Println("   ✓ R3: Repositories have no forbidden imports")
		fmt.Println("   ✓ R5: request_id middleware exists and is wired correctly")
	} else {
		fmt.Printf("❌ FAILED: %d error(s), %d warning(s)\n", len(r.Failures), len(r.Warnings))
		fmt.Println("   Fix all failures before committing.")
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
}
