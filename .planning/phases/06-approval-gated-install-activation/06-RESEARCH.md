# Phase 06: Approval-Gated Install & Activation - Research

**Researched:** 2026-03-10
**Domain:** Safe approval-gated installation and immediate activation of generated Codex skills in a Go CLI
**Confidence:** HIGH

## Summary

Phase 06 should be planned as the only mutating stage in the pipeline: it renders a final preview/diff, asks for explicit approval, performs an atomic write to `$CODEX_HOME/skills/<skill-name>`, and verifies post-install activation signals. This phase must consume Phase 05's explicit `ConflictResolutionDecision` and remain blocked if validation or conflict state is unresolved.

The recommended architecture is a fail-closed install transaction: `preflight gates -> preview/diff -> explicit confirm -> stage temp dir -> atomic move -> post-install verification`. For this project, no write should happen before a positive confirmation result in interactive mode, and non-interactive invocations should hard-fail unless an explicit force/approve flag is provided by policy.

Codex documentation indicates skill changes are detected automatically, with restart only as fallback when a change does not appear. Planning should treat "immediately usable" as "discoverable without extra setup in the normal path", and include deterministic verification checks plus a clear fallback message only for exceptional detection lag.

**Primary recommendation:** Implement Phase 06 as an atomic, fail-closed install transaction that requires explicit confirmation after a mandatory preview/diff and blocks on any unresolved Phase 04/05 gate state.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (`os`, `path/filepath`, `io/fs`) | Go 1.25.x stdlib | Path resolution, staging dirs, atomic rename, permission-safe writes | Canonical primitives for local filesystem transactions and install safety in Go CLIs. |
| `charm.land/huh/v2` | `v2.0.3` | Explicit yes/no approval prompts and deterministic interactive flow | Mature Go CLI prompt library with first-class `Confirm` and form primitives. |
| `github.com/sergi/go-diff` | `v1.4.0` | Human-readable pre-approval diff rendering for `SKILL.md` and touched files | Widely used text diff package with current 2025 release line and stable API. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/spf13/cobra` | `v1.10.2` | Command/flag mode control for interactive vs non-interactive approval policy | Use for explicit `RunE` error paths and deny-by-default non-interactive behavior. |
| `github.com/spf13/afero` | `v1.15.0` | Filesystem abstraction for install transaction tests | Use in unit/integration tests to verify fail-closed writes without mutating real `$CODEX_HOME`. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `go-diff` preview rendering | Hand-rolled line differ | Less dependency, but higher risk of misleading previews and edge-case bugs. |
| `huh` explicit prompts | Raw `stdin` parsing | Smaller stack, but weaker UX/input validation and harder to test reliably. |
| Single-step direct write | Stage + atomic move transaction | Direct write is simpler but violates safety goals under interruption/failure. |

**Installation:**
```bash
go get charm.land/huh/v2@v2.0.3
go get github.com/sergi/go-diff@v1.4.0
go get github.com/spf13/cobra@v1.10.2
go get github.com/spf13/afero@v1.15.0
```

## Architecture Patterns

### Recommended Project Structure
```text
internal/install/
├── model.go                   # install request/result + gate states
├── preflight_gates.go         # enforce VAL/OVLP resolution prerequisites
├── preview_diff.go            # render preview + unified/textual diff
├── approval_prompt.go         # explicit approval contract (interactive/non-interactive)
├── transaction.go             # stage -> write -> fsync/safe close -> atomic move
├── activate_verify.go         # post-install discoverability/usability checks
└── errors.go                  # typed fail-closed install errors

internal/app/generate/
└── install_stage.go           # consumes Phase 05 decision artifact and invokes install pipeline
```

### Pattern 1: Fail-Closed Preflight Gate
**What:** Block install unless Phase 04 validation is clear and Phase 05 conflict state is resolved and non-abort.
**When to use:** Always before preview or any filesystem preparation.
**Example:**
```go
func Preflight(state SessionState) error {
	if state.Validation.HasBlockingError() {
		return ErrInstallBlockedValidation
	}
	if !state.ConflictDecision.IsResolved() || state.ConflictDecision.Mode == ResolutionAbort {
		return ErrInstallBlockedConflict
	}
	return nil
}
```

### Pattern 2: Preview Before Approval, Never After
**What:** Show candidate content and/or diff first, then ask for explicit confirmation.
**When to use:** Every install/update/merge path (INST-01, INST-02).
**Example:**
```go
preview := RenderPreview(candidate, existingMaybe)
fmt.Println(preview)

var approved bool
_ = huh.NewConfirm().
	Title("Approve install to $CODEX_HOME/skills/<skill-name>?").
	Affirmative("Approve install").
	Negative("Cancel").
	Value(&approved).
	Run()
if !approved {
	return ErrInstallDeclined
}
```

### Pattern 3: Atomic Transactional Install
**What:** Write to temp staging under target parent, then `os.Rename` into final path after successful build/validation of staged artifact.
**When to use:** Any write/overwrite path to skill registry.
**Example:**
```go
stageDir, err := os.MkdirTemp(targetParent, ".skill-stage-*")
if err != nil { return err }
defer os.RemoveAll(stageDir)

if err := writeSkill(stageDir, artifact); err != nil { return err }
if err := verifyStaged(stageDir); err != nil { return err }

// same parent dir helps preserve atomic behavior expectations
if err := os.Rename(stageDir, finalPath); err != nil {
	return err
}
```

### Pattern 4: Activation Verification with Fallback Message
**What:** Verify installed skill is present and parse-valid immediately after write; surface restart fallback only if discovery lag occurs.
**When to use:** Always after successful transaction (INST-04).
**Example:**
```go
if err := VerifyInstalledSkill(finalPath); err != nil {
	return fmt.Errorf("installed but verification failed: %w", err)
}
fmt.Println("Skill installed and ready. If Codex does not show it yet, restart Codex.")
```

### Anti-Patterns to Avoid
- **Write then ask approval:** violates explicit approval requirement.
- **Proceed on unresolved conflict/validation warnings treated as acceptable:** breaks fail-closed policy.
- **In-place overwrite without staging:** risks partial/corrupt install state.
- **Silent non-interactive auto-approval:** unsafe in CI/scripts without explicit force semantics.
- **Treat "restart required" as default success path:** conflicts with immediate usability goal.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| CLI approval interaction | Custom rune-by-rune stdin parser | `huh` `Confirm`/`Select` | Better validation, accessibility options, deterministic behavior. |
| Text diff preview | Bespoke line-diff formatter | `go-diff` | Established diff algorithm support and clearer preview output. |
| Atomic install semantics | Ad hoc multi-step writes to final path | stdlib stage + `os.Rename` transaction | Safer failure behavior and easier reasoning/testing. |
| Gate orchestration | Boolean checks scattered in handlers | Centralized `Preflight` gate function | Prevents accidental bypass of fail-closed install contract. |

**Key insight:** Most Phase 06 defects come from sequencing errors (approval/write/gate order), not algorithm complexity; centralize gate + transaction flow and use proven prompt/diff/fs primitives.

## Common Pitfalls

### Pitfall 1: Approval prompt appears but write already happened
**What goes wrong:** User sees "approval" as cosmetic while filesystem has already mutated.
**Why it happens:** Preview/prompt logic is wired after install side effects.
**How to avoid:** Enforce operation order in a single orchestrator: `preflight -> preview -> confirm -> write`.
**Warning signs:** Declining prompt still leaves new/changed skill files on disk.

### Pitfall 2: Gate bypass via stale or missing Phase 05 decision
**What goes wrong:** Install proceeds even when conflict decision is unresolved or `abort`.
**Why it happens:** Decision artifact not required in install request type.
**How to avoid:** Make `ConflictResolutionDecision` required input and reject unresolved/abort states at preflight.
**Warning signs:** Install path succeeds with empty decision metadata.

### Pitfall 3: Non-atomic update corrupts existing skill
**What goes wrong:** Interrupted writes produce mixed old/new files.
**Why it happens:** In-place writes or cross-directory rename assumptions.
**How to avoid:** Stage in same parent, verify staged artifact, then single move into target.
**Warning signs:** Partial `SKILL.md` or orphaned temp files after interrupted run.

### Pitfall 4: Interactive assumptions break automation mode
**What goes wrong:** CI or piped runs hang at prompt or bypass approval unexpectedly.
**Why it happens:** No explicit TTY/non-TTY policy and no force-flag contract.
**How to avoid:** In non-interactive mode, fail closed unless explicit override flag is present and logged.
**Warning signs:** Jobs waiting on input, or installs occurring with no recorded explicit approval path.

### Pitfall 5: "Immediate usability" not verified
**What goes wrong:** Install succeeds but skill is not discoverable/usable and user gets no guidance.
**Why it happens:** No post-install verification step; assumption that write implies availability.
**How to avoid:** Verify presence + parse validity + registry visibility signal; emit fallback restart message only when needed.
**Warning signs:** Frequent user reports of "installed but not available."

## Code Examples

Verified patterns from official sources:

### Explicit Yes/No confirmation prompt
```go
// Source: https://pkg.go.dev/charm.land/huh/v2
var approved bool
err := huh.NewConfirm().
	Title("Are you sure?").
	Affirmative("Yes").
	Negative("No").
	Value(&approved).
	Run()
if err != nil {
	return err
}
```

### Diff generation before approval
```go
// Source: https://pkg.go.dev/github.com/sergi/go-diff
dmp := diffmatchpatch.New()
diffs := dmp.DiffMain(oldText, newText, false)
fmt.Println(dmp.DiffPrettyText(diffs))
```

### Atomic move caveat to account for in design
```go
// Source: https://pkg.go.dev/os#Rename
// "OS-specific restrictions may apply when oldpath and newpath are in different
// directories. Even within the same directory, on non-Unix platforms Rename is
// not an atomic operation."
if err := os.Rename(stagePath, finalPath); err != nil {
	return err
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Direct write then optional prompt | Mandatory preview + explicit approval before any write | Modern safety-first CLI workflows | Restores user trust and prevents accidental mutations. |
| Conflict checks decoupled from install gate | Typed preflight contract requiring resolved validation+conflict states | Current phase-driven gating architectures | Enforces fail-closed behavior across phase boundaries. |
| Assume restart required after install | Automatic detection first; restart as fallback | Current Codex skill docs | Supports INST-04 immediate usability goal in normal path. |

**Deprecated/outdated:**
- Treating confirmation as a UX nicety instead of a hard gate for writes.
- Allowing install progression with "best effort" unresolved conflict/validation state.

## Open Questions

1. **Definition of "immediately usable" test oracle**
   - What we know: Codex docs state automatic skill change detection; restart is fallback.
   - What's unclear: deterministic local check to prove implicit invocation readiness (beyond file presence/parse validity).
   - Recommendation: phase plan should define a practical acceptance oracle (e.g., install path verification + `/skills` visibility check in E2E harness).

2. **Non-interactive approval policy**
   - What we know: explicit approval is required before write.
   - What's unclear: exact CLI UX for automation (`--yes`, `--approve`, or reject all non-interactive usage).
   - Recommendation: set strict default to reject non-interactive install unless an explicit approval flag is passed and logged.

## Sources

### Primary (HIGH confidence)
- https://developers.openai.com/codex/skills - skill format, install locations, automatic detection, restart fallback language.
- https://pkg.go.dev/charm.land/huh/v2 - interactive prompt primitives including `Confirm`.
- https://pkg.go.dev/github.com/sergi/go-diff - diff/match/patch package details and usage.
- https://pkg.go.dev/os#Rename - official rename semantics and atomicity caveats.
- https://cobra.dev/docs/how-to-guides/working-with-commands/ - `RunE` error-driven command flow patterns for strict gate handling.

### Secondary (MEDIUM confidence)
- /Users/nickbohm/.codex/skills/.system/skill-installer/SKILL.md - local installer behavior (`$CODEX_HOME/skills` target and restart messaging convention).

### Tertiary (LOW confidence)
- Web ecosystem discovery queries for 2026 CLI confirmation and diff conventions were used for discovery only; critical claims above were verified with primary/official sources.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - libraries and APIs are directly documented in official package/docs sources.
- Architecture: HIGH - patterns directly implement INST-01/02/03/04 and locked fail-closed decisions.
- Pitfalls: MEDIUM-HIGH - derived from common CLI install sequencing failures, with strong alignment to official gate semantics.

**Research date:** 2026-03-10
**Valid until:** 2026-04-09
