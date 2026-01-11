package main

// ===============================================================
// ENFORCER RULES CONFIGURATION
// Customize these values for your repository
// ===============================================================

const (
	// Module name for import checking
	ModuleName = "ncoe"

	// Route prefix for HTMX fragment detection
	RoutePrefix = "/staff/"
)

// Paths to scan
var (
	TemplatePaths  = []string{"templates/**/*.html"}
	HandlerPaths   = []string{"internal/handler/*.go"}
	RepoPaths      = []string{"internal/repository/**/*.go"}
	MiddlewarePath = "internal/middleware"
	MainGoPath     = "cmd/server/main.go"
)

// R1: HTMX Fragment URL Rule
// All HTMX-triggered endpoints (hx-get, hx-post, etc.) that return fragments
// must include "/_" in the path to distinguish from full-page navigation.
var HTMXFragmentPatterns = []string{
	`hx-get="`,
	`hx-post="`,
	`hx-put="`,
	`hx-delete="`,
	`hx-patch="`,
}

// URLs that are allowed without "/_" (e.g., file downloads, full navigation)
var AllowedNonFragmentURLs = []string{
	// Add any URLs that should be exempt from the /_rule
	// Example: "/staff/documents/download",
}

// R2: Handler Forbidden Imports
// Handlers should only do HTTP I/O; they should NOT import these packages
var HandlerForbiddenImports = []string{
	"database/sql",
	ModuleName + "/internal/repository",
}

// R3: Repository Forbidden Imports
// Repositories should only do SQL; they should NOT import these packages
var RepoForbiddenImports = []string{
	ModuleName + "/internal/handler",
	ModuleName + "/internal/templates",
	ModuleName + "/templates",
	"net/http",
}

// R5: Request ID Requirements
var RequestIDRequirements = struct {
	MiddlewareFile string
	FunctionName   string
	HeaderName     string
}{
	MiddlewareFile: "internal/middleware/request_id.go",
	FunctionName:   "RequestID",
	HeaderName:     "X-Request-Id",
}

// Fragment route patterns in main.go that should use "/_"
// These suffixes indicate HTMX fragments (not full pages)
var FragmentRouteSuffixes = []string{
	"/panel",
	"/table",
	"/status",
	"/preview",
	"/modal",
	"/row",
	"/partial",
}
