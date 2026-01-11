# Claude Code Enforcer Instructions

When working on this codebase, always run the enforcer before and after making changes.

## Before Starting Work

```bash
./bin/enforce
```

Verify all checks pass before making changes.

## After Making Changes

```bash
./bin/enforce
```

Fix any failures before committing.

## Key Rules to Remember

1. **Fragment URLs use `/_`**: Any HTMX endpoint returning a partial must have `/_` in the path
   - `/staff/cases/{id}/_panel` (correct)
   - `/staff/cases/{id}/panel` (wrong)

2. **Layer boundaries**:
   - Handlers don't import `database/sql` or `internal/repository`
   - Repositories don't import `net/http` or `internal/handler`

3. **Request ID**: Every request has an `X-Request-Id` header for tracing

## Quick Reference

| Layer | Can Import | Cannot Import |
|-------|------------|---------------|
| Handler | service, templates | database/sql, repository |
| Service | repository, domain | net/http |
| Repository | database/sql, domain | net/http, handler, templates |

## If Enforcer Fails

Read the error message carefully. Common fixes:

- **R1 failure**: Add `/_` to HTMX fragment URLs in templates
- **R2 failure**: Move database logic to repository layer
- **R3 failure**: Move HTTP logic to handler layer
- **R5 failure**: Ensure request_id middleware is wired in main.go
