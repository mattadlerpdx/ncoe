# NCOE Enforcer

The enforcer is a code quality tool that validates architectural conventions and coding standards.

## Running the Enforcer

```bash
./bin/enforce
```

This runs three checks:
1. Go formatting (`gofmt`)
2. Unit tests (`go test`)
3. Architectural rules (`cmd/enforcer`)

## Exit Codes

- `0` - All checks passed
- `1` - One or more checks failed

## Rules

### R1: HTMX Fragment URLs (FAIL)

All HTMX endpoints that return partial HTML must include `/_` in the URL path.

**Correct:**
```html
<button hx-get="/staff/cases/{{.ID}}/_panel">View</button>
```

**Wrong:**
```html
<button hx-get="/staff/cases/{{.ID}}/panel">View</button>
```

**Fix:** Add `/_` before the fragment suffix (e.g., `/_panel`, `/_table`, `/_status`).

### R2: Handler Imports (FAIL)

Handlers (`internal/handler/*.go`) must not import:
- `database/sql`
- `ncoe/internal/repository`

**Why:** Handlers should only handle HTTP I/O. Database logic belongs in repositories.

**Fix:** Move database operations to a repository and call it via a service.

### R3: Repository Imports (FAIL)

Repositories (`internal/repository/**/*.go`) must not import:
- `net/http`
- `ncoe/internal/handler`
- `ncoe/internal/templates`

**Why:** Repositories should only handle data access. HTTP logic belongs in handlers.

**Fix:** Move HTTP-related code to the handler layer.

### R5: Request ID Middleware (FAIL)

The codebase must have:
1. File `internal/middleware/request_id.go`
2. Function `RequestID` in that file
3. Middleware wired in `cmd/server/main.go`
4. `X-Request-Id` header set on responses

**Why:** Request IDs enable distributed tracing and debugging.

**Fix:** Create the middleware file with the required function and wire it in main.go.

## Common Failure Examples

### R1 Failure

```
❌ FAILURES (1)
--------------------------------------------------

[R1] templates/staff/cases.html:166
  Message: HTMX fragment URL missing '/_': /staff/cases/{{.ID}}/panel
  Fix: Change to include '/_' before fragment suffix
```

**Solution:** Edit the template and change `/panel` to `/_panel`.

### R2 Failure

```
[R2] internal/handler/staff.go:15
  Message: Handler imports forbidden package: database/sql
  Fix: Handlers should only do HTTP I/O. Move business logic to services.
```

**Solution:** Remove the database import and call a service method instead.

## Customizing Rules

Edit `cmd/enforcer/rules.go` to:
- Add URLs to the allowlist (`AllowedNonFragmentURLs`)
- Add/remove forbidden imports
- Change the module name

## Extending the Enforcer

To add new rules:

1. Define rule configuration in `rules.go`
2. Add scanning logic in `scan.go`
3. Add check function in `main.go`
4. Update report output in `report.go`
