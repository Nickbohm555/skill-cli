# Testing Patterns

**Analysis Date:** 2026-03-10

## Test Framework

**Runner:**
- Go `testing` package (stdlib).
- Config: `go.mod` (no dedicated test runner config file such as `jest.config.*` or `vitest.config.*`).

**Assertion Library:**
- Standard `testing` assertions with `t.Fatalf`/`t.Fatal` and helper functions; no third-party assertion library is used.

**Run Commands:**
```bash
go test ./...              # Run all tests
go test ./...              # Watch mode (no dedicated watch mode configured)
go test -cover ./...       # Coverage
```

## Test File Organization

**Location:**
- Tests are co-located with implementation in the same package directories, for example `internal/crawl/engine.go` with `internal/crawl/engine_test.go`.

**Naming:**
- Use `*_test.go` naming consistently across packages, including `internal/install/transaction_test.go`, `internal/refinement/flow_test.go`, and `internal/content/extract_normalize_test.go`.

**Structure:**
```
cmd/cli-skill/
internal/<domain>/
  *.go
  *_test.go
```

## Test Structure

**Suite Organization:**
```go
func TestClassifyCandidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		// inputs + expected values
	}{
		{name: "docs article is accepted"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// execute + assert
		})
	}
}
```

**Patterns:**
- Setup pattern: construct fixtures with local variables and small helpers (`transactionTestRequest`, `readyAnswers`, `fieldWithAnswer`) in `internal/install/transaction_test.go`, `internal/refinement/flow_test.go`, and `internal/cli/prompts/refinement_form_test.go`.
- Teardown pattern: rely on `t.TempDir()` cleanup and `defer` server closes (`httptest.NewServer`) in `internal/install/transaction_test.go` and `internal/crawl/engine_test.go`.
- Assertion pattern: fail-fast checks with detailed failure context using `t.Fatalf`, including actual vs expected values.

## Mocking

**Framework:** Handwritten stubs/fakes and injected function seams.

**Patterns:**
```go
type stubQuestionAsker struct {
	primaryAnswers map[FieldID][]string
}

func (s *stubQuestionAsker) AskPrimary(field FieldState) (string, error) {
	return shiftAnswer(s.primaryAnswers, field.Definition.ID), nil
}
```

```go
type renameFailFS struct {
	transactionFS
	failTarget string
}

func (fs renameFailFS) Rename(oldpath, newpath string) error {
	if newpath == fs.failTarget {
		return errors.New("injected rename failure")
	}
	return fs.transactionFS.Rename(oldpath, newpath)
}
```

**What to Mock:**
- Mock boundary dependencies with package-private interfaces or injectable funcs, for example `transactionFS` in `internal/install/transaction.go` exercised by `internal/install/transaction_test.go`.
- Use test doubles for interaction flows and sequencing verification (`stubQuestionAsker`, `stubSummarizeFirstHandler`) in `internal/refinement/flow_test.go`.

**What NOT to Mock:**
- Do not mock pure internal logic that is deterministic and cheap to execute (URL normalization/classification, parsing, state transitions) in `internal/crawl/normalize_test.go`, `internal/crawl/classify_test.go`, and `internal/validation/validation_test.go`.

## Fixtures and Factories

**Test Data:**
```go
func transactionTestRequest(rootDir string) InstallRequest {
	request := previewTestRequest()
	request.Target = InstallTarget{RootDir: rootDir}
	request.Approval = explicitApprovedDecision()
	return request
}
```

**Location:**
- Fixtures are usually local helper functions at the bottom of the corresponding test file (`internal/install/transaction_test.go`, `internal/refinement/flow_test.go`, `internal/cli/prompts/refinement_form_test.go`).

## Coverage

**Requirements:** None enforced by config/CI artifacts detected in repository root.

**View Coverage:**
```bash
go test -cover ./...
```

## Test Types

**Unit Tests:**
- Dominant pattern. Most tests validate one function/module contract with table-driven cases (`internal/crawl/classify_test.go`, `internal/content/extract_normalize_test.go`, `internal/validation/validation_test.go`).

**Integration Tests:**
- Lightweight integration tests exist for network and filesystem boundaries:
- `internal/crawl/engine_test.go` uses `httptest.NewServer` to validate crawl behavior and skip taxonomy end-to-end.
- `internal/install/transaction_test.go` uses real filesystem interactions in `t.TempDir()` to validate atomic install/rollback behavior.

**E2E Tests:**
- Not used as a separate framework; no browser/system-level E2E suite detected.

## Common Patterns

**Async Testing:**
```go
func TestNormalizeEntryURL(t *testing.T) {
	t.Parallel()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
		})
	}
}
```

**Error Testing:**
```go
result, err := InstallTransaction(request)
if !errors.Is(err, ErrInstallApprovalRequired) {
	t.Fatalf("InstallTransaction() error = %v, want %v", err, ErrInstallApprovalRequired)
}
if !strings.Contains(err.Error(), "activate staged install") {
	t.Fatalf("Install() error = %v, want activation failure context", err)
}
```

---

*Testing analysis: 2026-03-10*
