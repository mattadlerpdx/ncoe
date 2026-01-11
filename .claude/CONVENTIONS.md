# NCOE Codebase Conventions

This document defines the architectural conventions and coding standards for the NCOE (Nevada Commission on Ethics) Case Management System.

## Philosophy

**"Uniform contracts + thin edges"**

- **Handlers**: HTTP I/O only (parse requests, call services, render responses)
- **Services**: Business logic + authorization
- **Repositories**: SQL/data access only
- **Templates**: Receive view models only (no business logic)
- **Fragments**: Explicit via routing contract

---

## Fragment URL Convention (CRITICAL)

All HTMX-triggered endpoints that return partial HTML fragments **MUST** include `/_` in the path.

### Rule

```
Fragment endpoints: include "/_" in path
Page endpoints: do NOT include "/_"
```

### Examples

| Type | URL | Description |
|------|-----|-------------|
| Page | `/staff/cases` | Full page with layout |
| Page | `/staff/cases/abc123` | Case detail full page |
| Fragment | `/staff/cases/abc123/_panel` | Case panel (HTMX partial) |
| Fragment | `/staff/cases/_table` | Cases table (HTMX partial) |
| Fragment | `/staff/cases/abc123/_status` | Status update action |

### Why?

1. **Explicit contract**: Anyone reading the code knows if a route returns a fragment
2. **Enforceable**: The enforcer tool can verify compliance
3. **Debuggable**: Network inspector shows fragment vs page requests clearly
4. **Prevents bugs**: Can't accidentally navigate to a fragment URL

### Template Usage

```html
<!-- CORRECT: Fragment URL with /_ -->
<button hx-get="/staff/cases/{{.ID}}/_panel"
        hx-target="#panelBody">View</button>

<!-- WRONG: Missing /_ for fragment -->
<button hx-get="/staff/cases/{{.ID}}/panel" ...>View</button>
```

---

## Request ID Tracking

Every HTTP request has a unique `X-Request-Id` header for tracing.

### How it works

1. `RequestID` middleware generates or propagates request IDs
2. ID is stored in request context
3. ID is added to response header `X-Request-Id`
4. Logging middleware includes the ID in all log lines

### Log Format

```
REQ=a1b2c3d4e5f6 GET /staff/cases 200 12.3ms user@example.com
```

### Accessing Request ID

```go
import "ncoe/internal/middleware"

func MyHandler(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    // Use for logging, error tracking, etc.
}
```

---

## Layer Boundaries

### Handlers (`internal/handler/`)

**DO:**
- Parse HTTP requests
- Validate input format
- Call service methods
- Render templates or JSON responses
- Handle HTTP-specific concerns (redirects, status codes)

**DON'T:**
- Import `database/sql`
- Import `internal/repository`
- Contain business logic
- Query the database directly

### Services (`internal/service/`)

**DO:**
- Implement business logic
- Check authorization
- Coordinate between repositories
- Return domain types

**DON'T:**
- Import `net/http`
- Handle HTTP requests
- Render templates

### Repositories (`internal/repository/`)

**DO:**
- Execute SQL queries
- Map database rows to domain types
- Handle transactions

**DON'T:**
- Import `net/http`
- Import `internal/handler`
- Import `internal/templates`
- Contain business logic

---

## HTMX Form Contract

For forms submitted via HTMX:

1. Use `hx-post` with a fragment URL (includes `/_`)
2. Server returns appropriate response:
   - Success: Return empty body + `HX-Trigger` header to refresh data
   - Validation error: Return form HTML with error messages
   - Redirect: Return empty body + `HX-Redirect` header

### Example

```html
<form hx-post="/staff/cases/{{.ID}}/_status" hx-swap="none">
    <select name="status">...</select>
    <button type="submit">Update</button>
</form>
```

```go
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request, id string) {
    // ... update logic ...

    // Trigger refresh
    w.Header().Set("HX-Trigger", "caseUpdated")
    w.WriteHeader(http.StatusOK)
}
```

---

## Running the Enforcer

Before committing, run:

```bash
./bin/enforce
```

This checks:
- R1: All HTMX fragment URLs use `/_` convention
- R2: Handlers have no forbidden imports
- R3: Repositories have no forbidden imports
- R5: request_id middleware exists and is wired

All FAIL checks must be fixed before committing.
