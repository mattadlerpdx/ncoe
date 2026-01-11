# Testing Guidelines

## Integration Test Guardrails

1. **Integration tests live in `/tests/integration/`**
   - Keep integration tests separate from unit tests
   - Unit tests stay in their respective packages

2. **All integration tests must use `internal/testutil.NewTestServer`**
   - Do not create ad-hoc test servers
   - The harness provides consistent setup and repository access

3. **No test may call `os.Chdir` or depend on working directory**
   - Use `runtime.Caller` to find project root if needed
   - Template paths must be absolute

4. **No `time.Sleep` in tests**
   - Use channels, waitgroups, or proper synchronization
   - If testing timeouts, use short test-specific values

5. **HTML must be parsed; no regex HTML**
   - Use `testutil.ParseDOM` for all HTML assertions
   - DOM helpers: `AssertFullPage`, `AssertFragment`, `AssertContainsText`, etc.

6. **Integration suite must run in <5 seconds locally**
   - Current benchmark: ~3 seconds for 51 tests
   - If tests slow down, investigate before merging

7. **Navigation integrity: UI links must resolve to valid data**
   - Dashboard links must point to cases that exist in the repository
   - Deadline references must point to cases with actual due dates
   - Any hardcoded display data must match repository state
   - See `TestNavigationIntegrity` in `flows_test.go`

## Running Tests

```bash
# Run all integration tests
go test -v ./tests/integration/...

# Run specific test
go test -v ./tests/integration/... -run TestAuthenticationFlow

# Run with race detector
go test -race ./tests/integration/...

# Run from any directory (VS Code, CI, etc.)
go test -v ./tests/integration/...
```

## Test Structure

```
ncoe/
├── internal/testutil/
│   ├── server.go    # TestServer harness, HTTP helpers
│   ├── dom.go       # DOM parsing and assertions
│   └── fixtures.go  # Test form data factories
│
└── tests/integration/
    ├── pages_test.go  # Table-driven route tests
    └── flows_test.go  # End-to-end workflow tests
```

## Data Validation

Tests should verify data is actually persisted, not just HTTP responses:

```go
// Good: Verify repository state
beforeCount := len(ts.Repos.Case.List("", "", ""))
ts.POST("/submit/advisory-opinion", form)
afterCount := len(ts.Repos.Case.List("", "", ""))
if afterCount != beforeCount+1 {
    t.Error("case was not created")
}

// Bad: Only check HTTP response
resp := ts.POST("/submit/advisory-opinion", form)
if resp.StatusCode != 303 {
    t.Error("expected redirect")
}
// No verification that case exists!
```

## Adding New Tests

1. For new routes: Add to `PublicPages` or `ProtectedPages` in `pages_test.go`
2. For new workflows: Add test function in `flows_test.go`
3. For new form types: Add fixture function in `fixtures.go`
