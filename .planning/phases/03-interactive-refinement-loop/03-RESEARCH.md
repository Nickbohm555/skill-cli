# Phase 3: Interactive Refinement Loop - Research

**Researched:** 2026-03-10
**Domain:** Go CLI adaptive conversation and revision loop
**Confidence:** HIGH

## Summary

Phase 03 should be planned as a deterministic conversation engine layered on top of the existing CLI command surface, not as ad-hoc prompt branching. The locked decisions in `03-CONTEXT.md` (domain-chunk questioning, confidence-driven deepening, `revise <field>`, sectioned review, and commit blocking) map cleanly to a small state machine plus a field dependency graph.

The strongest current stack for this phase is `cobra` for command orchestration plus `huh/v2` for interactive prompts and multiple-choice clarifications. `survey/v2` is explicitly no longer maintained, so planning should avoid it. Validation and readiness gating should be schema-backed (`jsonschema/v6`) plus deterministic clarity scoring rules, so "ready vs needs attention" is reproducible and testable.

Primary implementation strategy: run an iterative loop of `ask -> score clarity -> deepen or accept -> review -> revise impacts -> re-evaluate`, and only allow commit when every required field is both present and above clarity threshold.

**Primary recommendation:** Build Phase 03 around a typed `SessionState` + `FieldGraph` engine with `huh` prompts and schema-backed readiness gates; treat revision as first-class state transition, not a side path.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/spf13/cobra` | `v1.10.2` | Command lifecycle, `RunE`, and error/exit handling | Widely adopted CLI framework in Go; stable command architecture patterns |
| `charm.land/huh/v2` | `v2.0.3` | Interactive terminal prompts, select/multiselect, validation, accessibility mode | Actively maintained prompt/form library; supports dynamic forms and typed options |
| `github.com/santhosh-tekuri/jsonschema/v6` | `v6.0.2` | Readiness/required-field schema enforcement before commit | Strong JSON Schema compliance and mature validator surface |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/looplab/fsm` | `v1.0.3` | Explicit conversation state transitions | Use when flow branching complexity exceeds simple enum/switch loop |
| Go stdlib (`bufio`, `strings`, `regexp`) | Go toolchain | Parse `revise <field>` command and normalize free text | Use for revision command parsing and answer preprocessing |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `huh/v2` | `survey/v2` | `survey/v2` is no longer maintained; increases maintenance risk |
| `looplab/fsm` | hand-rolled switch state machine | Hand-rolled is fine for small flows; use `fsm` when transitions/events proliferate |
| `jsonschema/v6` | custom struct-tag validators only | Custom-only validation drifts and misses nested/conditional rule coverage |

**Installation:**
```bash
go get github.com/spf13/cobra@latest
go get charm.land/huh/v2@latest
go get github.com/santhosh-tekuri/jsonschema/v6@latest
go get github.com/looplab/fsm@latest
```

## Architecture Patterns

### Recommended Project Structure
```text
internal/
├── refinement/               # phase 03 package boundary
│   ├── session.go            # SessionState, required fields, readiness map
│   ├── flow.go               # state machine / loop orchestration
│   ├── question_policy.go    # confidence-driven deepening rules
│   ├── revise.go             # revise command parsing + impacted field expansion
│   ├── review.go             # sectioned review renderer + readiness badges
│   └── validator.go          # required/clarity/schema gate
└── cli/
    └── prompts.go            # huh adapters (Input/Select/Confirm wrappers)
```

### Pattern 1: Domain-Chunk Adaptive Loop
**What:** Ask by section (`purpose`, `constraints`, `examples`, `boundaries`) and only deepen fields that are missing or low clarity.
**When to use:** Default phase interaction path for INT-01 and INT-02.
**Example:**
```go
// Source: https://pkg.go.dev/charm.land/huh/v2
for _, section := range orderedSections {
    for _, field := range requiredBySection(section) {
        if !state.IsReady(field) {
            askPrimaryPrompt(field)
            score := clarity.Score(state.Answer(field))
            if score < thresholds[field] {
                askTargetedClarification(field) // usually Select + "Other"
            }
        }
    }
}
```

### Pattern 2: Revision as Dependency-Aware Revalidation
**What:** `revise <field>` updates one answer, then invalidates and re-asks impacted follow-ups using a dependency graph.
**When to use:** Any pre-commit edit request (INT-03).
**Example:**
```go
// Source: project pattern derived from phase decisions
func ReviseField(field string, state *SessionState, graph FieldGraph) error {
    newValue := askRevisionPrompt(field)
    state.Set(field, newValue)

    impacted := graph.ImpactedBy(field)
    for _, f := range impacted {
        state.MarkNeedsAttention(f)
        askTargetedClarification(f)
    }
    return nil
}
```

### Pattern 3: Sectioned Final Review with Commit Gate
**What:** Render grouped review and block commit when required fields are missing or below clarity floor.
**When to use:** Before generation handoff.
**Example:**
```go
// Source: https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6
report := validator.Evaluate(state)
renderSectionedReview(report) // ready / needs-attention per field

if !report.CommitReady {
    return errors.New("cannot commit: required fields missing or low clarity")
}
```

### Anti-Patterns to Avoid
- **Flat one-pass questionnaire:** violates adaptive deepening and causes over/under-questioning.
- **Revision without dependency invalidation:** creates stale derived answers and false readiness.
- **Commit on "required present" only:** ignores clarity quality bar from phase decision.
- **Prompt logic as policy source of truth:** keep readiness/clarity logic in deterministic domain code.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Interactive terminal widgets | Custom arrow-key/select renderer | `huh/v2` fields (`Select`, `Input`, `Confirm`) | Terminal behavior, accessibility, and rendering edge cases are already handled |
| Schema-level commit gate | Ad-hoc nested `if` checks across files | `jsonschema/v6` + explicit readiness report | Centralized, auditable constraints; easier to evolve |
| Complex transition graph | Unstructured bool flags | `looplab/fsm` (or strict enum state machine) | Prevents illegal transitions and hidden dead-ends |

**Key insight:** hand-rolled interactive terminal controls and quality gates accumulate subtle edge-case debt quickly; use battle-tested libraries for IO/validation and keep domain policy custom.

## Common Pitfalls

### Pitfall 1: Infinite Deepening Loops
**What goes wrong:** Ambiguous answers trigger repeated clarifications without convergence.
**Why it happens:** No per-field attempt cap or fallback path.
**How to avoid:** Set max deepening attempts and switch to structured multiple choice + "other".
**Warning signs:** Same field asked >2 times with no readiness improvement.

### Pitfall 2: Readiness Drift After Revision
**What goes wrong:** User revises upstream field, but dependent fields stay marked ready.
**Why it happens:** No dependency graph invalidation.
**How to avoid:** Maintain explicit `FieldGraph` and invalidate impacted nodes on revise.
**Warning signs:** Final review shows "ready" fields inconsistent with revised intent.

### Pitfall 3: CLI UX Split-Brain
**What goes wrong:** Some prompts come from one style path and others from another, confusing users.
**Why it happens:** Mixing raw stdin prompts and form widgets inconsistently.
**How to avoid:** Keep a single prompt adapter surface and standardized prompt phrasing.
**Warning signs:** Inconsistent keybindings, confirmation semantics, or accessibility behavior.

### Pitfall 4: Commit Gate Too Weak
**What goes wrong:** Flow allows finalization with low-clarity required fields.
**Why it happens:** Gate checks only presence, not clarity score.
**How to avoid:** Gate on both completeness and per-field clarity threshold.
**Warning signs:** Frequent downstream validation failures in later phases.

## Code Examples

Verified patterns from official sources:

### Single-field typed prompt with validation
```go
// Source: https://pkg.go.dev/charm.land/huh/v2
var purpose string
err := huh.NewInput().
    Title("What is the skill's primary purpose?").
    Validate(func(s string) error {
        if strings.TrimSpace(s) == "" {
            return errors.New("purpose is required")
        }
        return nil
    }).
    Value(&purpose).
    Run()
```

### Targeted clarification with select + other
```go
// Source: https://pkg.go.dev/charm.land/huh/v2
var boundaryMode string
_ = huh.NewSelect[string]().
    Title("How strict should scope boundaries be?").
    Options(
        huh.NewOption("Strict single-capability", "strict"),
        huh.NewOption("Balanced", "balanced"),
        huh.NewOption("Other (describe)", "other"),
    ).
    Value(&boundaryMode).
    Run()
```

### Explicit transition handling
```go
// Source: https://pkg.go.dev/github.com/looplab/fsm
flow := fsm.NewFSM(
    "collecting",
    fsm.Events{
        {Name: "to_review", Src: []string{"collecting"}, Dst: "review"},
        {Name: "to_revise", Src: []string{"review"}, Dst: "collecting"},
        {Name: "commit", Src: []string{"review"}, Dst: "committed"},
    },
    fsm.Callbacks{},
)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Static question lists | Adaptive, confidence-driven questioning | Commonly adopted in modern assistant UX (2024+) | Fewer unnecessary prompts, higher answer quality |
| `survey/v2` for terminal prompts | `huh/v2` / Bubble Tea ecosystem | `survey` marked unmaintained; `huh` v2 active in 2025 | Better maintenance outlook and richer dynamic forms |
| Presence-only completion checks | Completeness + clarity readiness gates | Quality-gated generation workflows (recent CLI assistants) | Fewer low-quality outputs passed downstream |

**Deprecated/outdated:**
- `github.com/AlecAivazis/survey/v2`: marked "no longer maintained" in its package README; avoid for new phase implementation.

## Open Questions

1. **Clarity scoring formula (rule-based vs model-assisted)**
   - What we know: confidence-driven branching is required; threshold operationalization is discretionary.
   - What's unclear: exact scoring rubric and calibration target.
   - Recommendation: start with deterministic rules (length + specificity signals + ambiguity keyword penalties), then tune with fixture-based tests.

2. **Persistence scope of refinement session**
   - What we know: revision before commit is required within session.
   - What's unclear: whether resume across process restarts is in this phase or later.
   - Recommendation: plan in-memory for Phase 03 with serialization seam, then make persistence explicit in a later phase unless requested.

## Sources

### Primary (HIGH confidence)
- https://pkg.go.dev/github.com/spf13/cobra - command architecture and current module versions
- https://cobra.dev/docs/how-to-guides/working-with-commands/ - `RunE`, command organization, error handling patterns
- https://pkg.go.dev/charm.land/huh/v2 - interactive form APIs, validation, dynamic form capabilities
- https://github.com/charmbracelet/huh/releases - release activity and latest `v2.0.3` notes (2025-03-10)
- https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6 - schema validation capabilities and current version
- https://pkg.go.dev/github.com/looplab/fsm - state machine API and latest stable version
- https://pkg.go.dev/github.com/AlecAivazis/survey/v2 - unmaintained notice in README

### Secondary (MEDIUM confidence)
- https://github.com/charmbracelet/huh - ecosystem usage context and project activity signal

### Tertiary (LOW confidence)
- Web search summaries for "best" libraries and comparative claims (used only for discovery; recommendations verified via official docs above)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - based on official package docs and release pages
- Architecture: HIGH - derived directly from locked phase decisions plus validated library capabilities
- Pitfalls: MEDIUM-HIGH - pattern-derived, consistent with validated flow constraints

**Research date:** 2026-03-10
**Valid until:** 2026-04-09
