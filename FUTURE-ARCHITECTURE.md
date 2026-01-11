# NCOE Future Architecture Guide

> **Status**: This is a demo application. These are recommendations for future development if NCOE moves to production.

Based on patterns from the MyFlow project (Springfield Utility Board).

---

## Current State (Demo)

The current implementation works but has some technical debt:

1. **Mixed Template Patterns**: Dashboard is standalone, other pages use `staff_base`
2. **Renderer uses strategy heuristics**: Tries multiple patterns to find the right template
3. **No explicit Page/Fragment separation**: Panels render as pages wrapped in base

This is acceptable for a demo but should be addressed for production.

---

## Recommended Architecture (Production)

### 1. Template Convention

**Single Pattern for Pages:**
All staff pages should define `{{define "content"}}` and `{{define "title"}}` blocks.
The renderer always executes `staff_base` which includes these blocks.

```html
<!-- templates/staff/dashboard.html -->
{{define "title"}}Dashboard{{end}}

{{define "content"}}
    <!-- Page content here -->
{{end}}
```

**Fragments use `/_` URL convention:**
```
/staff/cases          → Full page
/staff/cases/_table   → Table fragment (HTMX)
/staff/cases/1/_panel → Detail panel fragment (HTMX)
```

### 2. Renderer Simplification

Replace strategy-based rendering with deterministic methods:

```go
// Renderer provides Page() and Fragment() methods
type Renderer struct {
    templates map[string]*template.Template
}

// Page renders a full page using the base template
func (r *Renderer) Page(w http.ResponseWriter, name string, data interface{}) error {
    tmpl := r.templates[name]
    return tmpl.ExecuteTemplate(w, "staff_base", data)
}

// Fragment renders an HTMX fragment directly (no base)
func (r *Renderer) Fragment(w http.ResponseWriter, name string, data interface{}) error {
    tmpl := r.templates[name]
    return tmpl.ExecuteTemplate(w, name, data)
}
```

### 3. Handler/Service/Repository Boundaries

**Handlers** (`internal/handler/`):
- Parse HTTP request
- Call service methods
- Render templates
- **Forbidden**: Direct DB access, business logic

**Services** (`internal/service/`):
- Business logic
- Authorization checks
- Orchestrate repository calls
- **Forbidden**: HTTP code, template rendering

**Repositories** (`internal/repository/`):
- SQL queries only
- Data mapping (rows → structs)
- **Forbidden**: HTTP code, business logic

### 4. Presenter Pattern (Optional for Larger Apps)

For production apps, consider the presenter pattern to keep templates "dumb":

```go
// internal/ui/case.go
type CaseRowData struct {
    ID           string
    CaseNumber   string
    TypeBadge    string  // "bg-primary", "bg-danger", etc.
    StatusBadge  string
    OverdueFlag  bool
    // Pre-computed display values only
}

func CaseRow(c *domain.Case) CaseRowData {
    return CaseRowData{
        ID:          c.ID,
        CaseNumber:  c.CaseNumber,
        TypeBadge:   typeToBadgeClass(c.Type),
        StatusBadge: statusToBadgeClass(c.Status),
        OverdueFlag: c.IsOverdue(),
    }
}
```

Templates then just read flags:
```html
{{if .OverdueFlag}}<span class="badge bg-danger">Overdue</span>{{end}}
```

### 5. Integration Testing Pattern

Two-layer assertions for every test:

1. **UI renders correctly** (no 500s, correct template wiring)
2. **Behavior is correct** (DB side-effects + visible UI changes)

```go
func TestCaseStatusUpdate(t *testing.T) {
    ts := testutil.NewTestServer(t)
    defer ts.Close()
    ts.Login("test@test.gov", "password")

    // Step 1: Guard - page MUST render
    ts.GuardPageOK("/staff/cases")

    // Step 2: Perform mutation
    ts.POST("/staff/cases/1/_status", url.Values{"status": {"under_review"}})

    // Step 3: Assert DB - verify side effects
    c := ts.Repos.Case.GetByID("1")
    if c.Status != "under_review" {
        t.Error("status not updated in repository")
    }

    // Step 4: Assert UI - verify visible change
    body := ts.GET("/staff/cases/1")
    if !strings.Contains(body, "In Review") {
        t.Error("status badge not showing in UI")
    }
}
```

---

## Migration Path

If moving to production, prioritize in this order:

### Phase 1: Template Normalization
1. Convert `dashboard.html` to use `content` block pattern
2. Simplify renderer to deterministic Page/Fragment methods
3. Update all handlers to use explicit methods

### Phase 2: Layer Separation
1. Move business logic from handlers to services
2. Ensure repositories have no business logic
3. Add service-level tests

### Phase 3: Testing Improvements
1. Add Golden tests for all pages/fragments
2. Add mutation tests with DB assertions
3. Add step-by-step logging to all tests

### Phase 4: Advanced Patterns (Optional)
1. Implement presenter pattern for complex pages
2. Add audit logging for all mutations
3. Add request ID correlation

---

## Reference

See MyFlow project for complete implementation examples:
- `C:\Users\matt\Desktop\SUB\MyFlow\.claude\CONVENTIONS.md`
- `C:\Users\matt\Desktop\SUB\MyFlow\.claude\architecture.md`
- `C:\Users\matt\Desktop\SUB\MyFlow\internal\render\render.go`

---

## Current Known Issues

1. **Panel templates wrapped in base**: `case_panel.html` and `acknowledgment_panel.html`
   are rendered through the base template strategy, making them full pages instead of
   true HTMX fragments. This works for the demo but should be fixed for production.

2. **Strategy-based rendering**: The renderer tries multiple patterns which can be
   unpredictable. Should be replaced with explicit Page/Fragment methods.

3. **Mixed template patterns**: Dashboard is standalone, others use base. Should
   normalize to single pattern.
