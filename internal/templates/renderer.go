package templates

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// Renderer handles template parsing and rendering
type Renderer struct {
	pages   map[string]*template.Template
	funcMap template.FuncMap
	quiet   bool
}

// NewRenderer creates a new template renderer loading templates from templateDir
func NewRenderer(templateDir string) *Renderer {
	return newRenderer(templateDir, false)
}

// NewQuietRenderer creates a renderer that suppresses logging (for tests)
func NewQuietRenderer(templateDir string) *Renderer {
	return newRenderer(templateDir, true)
}

func newRenderer(templateDir string, quiet bool) *Renderer {
	funcMap := template.FuncMap{
		"formatDate": func(t interface{}) string {
			return ""
		},
		"statusBadge": func(status string) string {
			return status
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
	}

	renderer := &Renderer{
		pages:   make(map[string]*template.Template),
		funcMap: funcMap,
		quiet:   quiet,
	}

	// Folders that have templates
	folders := []string{"auth", "public", "staff", "errors"}

	for _, folder := range folders {
		folderPath := filepath.Join(templateDir, folder)

		// Get base template for this folder (if exists)
		baseFile := filepath.Join(folderPath, "base.html")
		var baseFiles []string
		if matches, _ := filepath.Glob(baseFile); len(matches) > 0 {
			baseFiles = matches
		}

		// Get all partials for this folder (_*.html)
		partialPattern := filepath.Join(folderPath, "_*.html")
		partialFiles, _ := filepath.Glob(partialPattern)

		// Get all page templates (non-partial *.html)
		allPattern := filepath.Join(folderPath, "*.html")
		allFiles, _ := filepath.Glob(allPattern)

		var pageFiles []string
		for _, f := range allFiles {
			base := filepath.Base(f)
			if !strings.HasPrefix(base, "_") && base != "base.html" {
				pageFiles = append(pageFiles, f)
			}
		}

		// Parse each page template with base + partials
		for _, pageFile := range pageFiles {
			relPath, _ := filepath.Rel(templateDir, pageFile)
			name := strings.TrimSuffix(relPath, ".html")
			name = filepath.ToSlash(name)

			// Parse page with base + partials
			var files []string
			files = append(files, baseFiles...)
			files = append(files, partialFiles...)
			files = append(files, pageFile)

			tmpl, err := template.New(filepath.Base(pageFile)).Funcs(funcMap).ParseFiles(files...)
			if err != nil {
				if !quiet {
					log.Printf("Failed to parse page %s: %v", name, err)
				}
				continue
			}

			renderer.pages[name] = tmpl
			if !quiet {
				log.Printf("Loaded page: %s", name)
			}
		}
	}

	return renderer
}

// ExecuteTemplate renders a page template
func (r *Renderer) ExecuteTemplate(w http.ResponseWriter, name string, data interface{}) error {
	tmpl, ok := r.pages[name]
	if !ok {
		if !r.quiet {
			log.Printf("Page template not found: %s", name)
		}
		return http.ErrMissingFile
	}

	// Try different execution strategies:
	// 1. If page defines its own "folder/name.html", execute that (standalone pages like dashboard)
	// 2. Try just "name.html" (for templates like case_panel.html)
	// 3. If page uses base template pattern (staff_base), execute that (pages with content block)

	strategies := []string{
		name + ".html",                // Standalone page pattern (e.g., "staff/dashboard.html")
		filepath.Base(name) + ".html", // Just filename (e.g., "case_panel.html")
		"staff_base",                  // Base template pattern (most staff pages)
	}

	for _, strategy := range strategies {
		if t := tmpl.Lookup(strategy); t != nil {
			err := tmpl.ExecuteTemplate(w, strategy, data)
			if err != nil && !r.quiet {
				log.Printf("Page execution error (%s via %s): %v", name, strategy, err)
			}
			return err
		}
	}

	if !r.quiet {
		log.Printf("No valid template found for %s, tried: %v", name, strategies)
	}
	return http.ErrMissingFile
}
