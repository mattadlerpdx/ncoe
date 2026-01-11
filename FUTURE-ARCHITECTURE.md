# NCOE Future Architecture Guide

> **Status**: Demo application. Recommendations for production hardening (patterned after MyFlow / SUB).

---

## Current State (Demo)

- **Mixed template patterns**: some pages are standalone; others rely on `staff_base` + blocks.
- **Strategy-based template execution**: renderer tries multiple template names/paths.
- **Fragments not first-class**: HTMX panels sometimes render through base layout (full HTML document).

Acceptable for demo; too risky for production because it causes blank pages or wrapped fragments.

---

## Recommended Architecture (Production)

### 1) Template Conventions

#### Pages (single pattern)

All staff pages define only blocks:

```html
{{define "title"}}Dashboard{{end}}

{{define "content"}}
  <!-- page content -->
{{end}}
```

Pages are rendered by executing `staff_base` (which includes the blocks).

#### Fragments (HTMX)

Fragments use a route naming convention and render without base layout:

```
/staff/cases          → full page
/staff/cases/_table   → fragment
/staff/cases/1/_panel → fragment
```

Fragment templates:
- Live in `_*.html` files
- Define a unique template name like `staff/_case_panel`
- Never include `<html>`, `<head>`, `<body>`, or call `staff_base`

Example:

```html
{{define "staff/_case_panel"}}
<div id="case-panel">
  ...
</div>
{{end}}
```

**Note**: `/_` is a routing convention only. Handlers must explicitly call `render.Page()` or `render.Fragment()`.

---

### 2) Renderer Simplification (Deterministic)

Remove all heuristic "try patterns" logic.

Use explicit methods:

```go
type Renderer struct {
    staff *template.Template // staff template set (includes staff_base + all staff pages + fragments)
}

func (r *Renderer) Page(w http.ResponseWriter, data any) error {
    return r.staff.ExecuteTemplate(w, "staff_base", data)
}

func (r *Renderer) Fragment(w http.ResponseWriter, fragmentName string, data any) error {
    return r.staff.ExecuteTemplate(w, fragmentName, data)
}
```

- `Page()` always executes `staff_base`.
- `Fragment()` always executes the fragment define name (e.g. `staff/_case_panel`).

---

### 3) Boundaries: Handler / Service / Repository

#### Handlers (`internal/handler/`)
- Parse HTTP request
- Call service methods
- Choose `Page` vs `Fragment` render
- **Forbidden**: direct DB access, business rules

#### Services (`internal/service/`)
- Business logic + authorization
- Orchestrate repositories
- **Forbidden**: HTTP concerns, template rendering

#### Repositories (`internal/repository/`)
- SQL + row mapping only
- **Forbidden**: HTTP concerns, business logic

---

### 4) Presenter Pattern (Optional)

For larger UI, use presenters so templates only read precomputed view fields.

```go
type CaseRowVM struct {
    ID          string
    CaseNumber  string
    TypeClass   string
    StatusLabel string
    Overdue     bool  // computed in presenter, not via domain method
}
```

Templates render simple flags:

```html
{{if .Overdue}}<span class="badge bg-danger">Overdue</span>{{end}}
```

---

### 5) Integration Testing Pattern

Every integration test has two layers:

1. **Rendering guardrail** (no 500s, correct page/fragment shape)
2. **Behavior correctness** (state changes + visible UI result)

```go
ts.GuardPageOK("/staff/cases")            // asserts <html> present
ts.GuardFragmentOK("/staff/cases/_table") // asserts <html> absent
```

Then behavior assertions:
- DB/repo side effects
- UI text/DOM changes

---

## Migration Path

### Phase 1: Template Normalization
1. Convert all staff pages to the block pattern (`title`/`content`)
2. Remove strategy-based renderer logic
3. Standardize fragment templates to `_*.html` + `staff/_name`

### Phase 2: Layer Separation
1. Move business logic into services
2. Keep repositories SQL-only

### Phase 3: Testing Hardening
1. Golden render tests for all pages/fragments
2. Mutation tests with DB + UI assertions

### Phase 4 (Optional)
- Presenter pattern
- Audit logging
- Request IDs

---

## Current Known Issues (Demo)

1. Some panels render via base layout (not true fragments)
2. Renderer uses strategy-based template guessing
3. Mixed page patterns (standalone vs base-block)

---

## Reference

See MyFlow project for complete implementation examples:
- `C:\Users\matt\Desktop\SUB\MyFlow\.claude\CONVENTIONS.md`
- `C:\Users\matt\Desktop\SUB\MyFlow\.claude\architecture.md`
- `C:\Users\matt\Desktop\SUB\MyFlow\internal\render\render.go`
