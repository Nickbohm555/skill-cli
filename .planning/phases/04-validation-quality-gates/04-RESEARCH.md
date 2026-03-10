# Phase 04: Validation & Quality Gates - Research

**Researched:** 2026-03-10
**Domain:** Strict validation gates for generated Codex skills in Go CLI
**Confidence:** HIGH

## Summary

Phase 04 should be planned as a deterministic validation subsystem that sits between generation and any downstream progression/install step. The validator must fail closed on structural violations, block progression on blocking issues, and drive a guided single-issue fix loop until the candidate is acceptable or the user exits.

For this project, the strongest implementation is a layered validator: parse `SKILL.md` into a normalized model, run strict structural/schema checks first, then run semantic quality checks (including explicit in-scope and out-of-scope boundaries). Validation output should be machine-readable with stable rule IDs and severities (`Error`, `Warning`) so Phase 03's interaction loop can generate targeted follow-up prompts and revalidate one edit at a time.

This phase should not attempt broad UX redesign or install flow logic; it should define the quality gate contract used by later phases. The main planning risk is ambiguity around "semantic acceptability", so tasks should explicitly codify rule definitions, failure messages, and prompt mapping per rule.

**Primary recommendation:** Implement a two-pass validator (`structural -> semantic`) with rule IDs, fail-closed blocking on any `Error`, and a one-issue-at-a-time targeted follow-up loop that revalidates after each single edit.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/yuin/goldmark` | `v1.7.16` | Parse Markdown body and heading structure from `SKILL.md` | CommonMark-compliant AST parser with explicit extension points for robust section validation. |
| `go.abhg.dev/goldmark/frontmatter` | `v0.3.0` | Parse and decode YAML/TOML frontmatter into typed structs | Purpose-built frontmatter support integrated with goldmark parser context. |
| `github.com/santhosh-tekuri/jsonschema/v6` | `v6.0.2` | Strict schema validation for normalized skill model | Strong JSON Schema compliance (incl. draft 2020-12) with introspectable errors suited for targeted prompts. |
| `github.com/go-playground/validator/v10` | `v10.30.1` | Deterministic semantic/rule-level validation beyond schema | Mature rule/tag + custom-validator pattern for cross-field semantics and stable error extraction. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `charm.land/huh/v2` | `v2.0.3` | One-question targeted follow-up prompts after validation failures | Use in retry loop to present exactly one fix request and revalidate. |
| `gopkg.in/yaml.v3` | `v3.0.1` | YAML decoding helpers for frontmatter and metadata handling | Use when explicit YAML-node control is needed outside parser extension defaults. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `go.abhg.dev/goldmark/frontmatter` + `goldmark` | `github.com/adrg/frontmatter` + custom Markdown parsing | Simpler frontmatter-only parse, but requires separate body/heading parser orchestration. |
| `jsonschema/v6` for structural contracts | Hand-rolled map/struct checks only | Faster initial coding, but weaker error consistency and higher long-term rule drift risk. |
| `validator/v10` for semantic cross-field checks | All semantic checks as ad hoc `if` statements | Less dependency surface, but poorer reuse/composability and harder standardized error metadata. |

**Installation:**
```bash
go get github.com/yuin/goldmark@v1.7.16
go get go.abhg.dev/goldmark/frontmatter@v0.3.0
go get github.com/santhosh-tekuri/jsonschema/v6@v6.0.2
go get github.com/go-playground/validator/v10@v10.30.1
go get charm.land/huh/v2@v2.0.3
go get gopkg.in/yaml.v3@v3.0.1
```

## Architecture Patterns

### Recommended Project Structure
```text
internal/validation/
├── model.go                # normalized candidate model from SKILL.md
├── parse_skill.go          # frontmatter + markdown parsing
├── schema_validate.go      # structural/schema checks (VAL-01)
├── semantic_validate.go    # boundary + quality checks (VAL-03)
├── report.go               # issue model, severity, rule IDs, ordering
└── followup_prompt.go      # issue->targeted prompt mapping (VAL-02)

internal/app/generate/
└── fix_loop.go             # one-issue-at-a-time retry orchestration
```

### Pattern 1: Two-Pass Fail-Closed Validation
**What:** Run structural/schema checks first, semantic checks second; any `Error` blocks progression.
**When to use:** Always for Phase 04 acceptance decisions.
**Example:**
```go
// Source: https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6
// plus project fail-closed policy.
report := validation.NewReport()

structIssues := validator.Structural(candidate)
report.Add(structIssues...)
if report.HasError() {
    return report // block immediately; no progression
}

semanticIssues := validator.Semantic(candidate)
report.Add(semanticIssues...)
return report // caller decides pass/block from severity
```

### Pattern 2: Stable Rule IDs for Prompt Targeting
**What:** Each validation failure has a stable `rule_id` and severity to map into one specific follow-up prompt.
**When to use:** Required for VAL-02 targeted re-ask behavior.
**Example:**
```go
type ValidationIssue struct {
    RuleID   string   // e.g. "VAL.SECTION.IN_SCOPE_MISSING"
    Severity Severity // Error | Warning
    Message  string
    Path     string   // e.g. "sections.in_scope"
}

func NextBlockingIssue(r Report) (ValidationIssue, bool) {
    for _, issue := range r.Issues {
        if issue.Severity == Error {
            return issue, true // one-at-a-time guidance
        }
    }
    return ValidationIssue{}, false
}
```

### Pattern 3: Guided Single-Edit Revalidation Loop
**What:** Present one blocking issue, ask one targeted follow-up, apply one edit, revalidate immediately.
**When to use:** Every failure cycle (per locked context decisions).
**Example:**
```go
for {
    report := validate(candidate)
    issue, ok := report.NextBlockingIssue()
    if !ok {
        break // no blocking errors
    }

    prompt := followup.ForRule(issue.RuleID)
    answer := askOne(prompt) // one issue, one prompt
    candidate = applySingleEdit(candidate, issue, answer)
}
```

### Anti-Patterns to Avoid
- **Batching all issues into one mega prompt:** breaks one-at-a-time decision and reduces fix precision.
- **Severity-free validation output:** prevents deterministic block/pass behavior.
- **Implicit "best effort" acceptance:** violates fail-closed VAL-01 requirement.
- **Regex-only section checks on raw markdown:** brittle; parse to structured model first.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Markdown + heading parsing | Custom line-scanner parser | `goldmark` AST parsing | Section detection and heading semantics are edge-case heavy. |
| Frontmatter extraction | Manual delimiter slicing | `goldmark/frontmatter` | Handles typed decode and parser-context integration cleanly. |
| Schema conformance engine | Custom required-field checker only | `jsonschema/v6` | Better compliance, richer error details, less drift over time. |
| Cross-field semantic checks | Scattered `if` chains in handlers | `validator/v10` + explicit custom rules | Keeps semantic policy centralized and testable. |

**Key insight:** In this phase, reliability comes from deterministic parsing and standardized validation primitives; hand-rolled parsers/checkers create hidden acceptance bugs that directly undermine fail-closed gating.

## Common Pitfalls

### Pitfall 1: Structural pass but semantic failure is ignored
**What goes wrong:** Candidate passes schema but still has vague or missing scope boundaries.
**Why it happens:** Teams stop at field-presence validation.
**How to avoid:** Make semantic validation a required second pass, not optional linting.
**Warning signs:** Many accepted skills have generic "does everything" descriptions.

### Pitfall 2: Non-deterministic issue ordering
**What goes wrong:** Different runs surface different "first" issues, confusing the fix loop.
**Why it happens:** Issues are emitted from unordered maps or parallel checks without sorting.
**How to avoid:** Sort by deterministic priority (`Error` first, then rule weight, then path).
**Warning signs:** Same input yields different first follow-up prompt across runs.

### Pitfall 3: Warning severity leakage into blocking path
**What goes wrong:** Non-blocking `Warning` items accidentally block progression.
**Why it happens:** Gate logic checks for "any issues" instead of "any Error".
**How to avoid:** Centralize `HasBlockingIssues()` and use that everywhere.
**Warning signs:** User gets blocked despite no `Error` in report.

### Pitfall 4: Follow-up prompts are generic, not issue-targeted
**What goes wrong:** Retry loop stalls because user gets broad prompts not tied to failed rule.
**Why it happens:** No rule-to-prompt mapping table.
**How to avoid:** Create explicit mapping from `rule_id -> prompt template -> expected edit target`.
**Warning signs:** Multiple retries without changing the same failing rule outcome.

## Code Examples

Verified patterns from official sources:

### Parse frontmatter with goldmark extension
```go
// Source: https://pkg.go.dev/go.abhg.dev/goldmark/frontmatter
md := goldmark.New(
    goldmark.WithExtensions(&frontmatter.Extender{}),
)
ctx := parser.NewContext()
_ = md.Convert(src, io.Discard, parser.WithContext(ctx))

fm := frontmatter.Get(ctx)
var meta struct {
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
}
if err := fm.Decode(&meta); err != nil {
    return err
}
```

### Compile and apply JSON Schema validation
```go
// Source: https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6
compiler := jsonschema.NewCompiler()
if err := compiler.AddResource("skill.schema.json", strings.NewReader(schemaJSON)); err != nil {
    return err
}
schema, err := compiler.Compile("skill.schema.json")
if err != nil {
    return err
}
if err := schema.Validate(normalizedSkill); err != nil {
    return err // map to Error severity in report
}
```

### Field and cross-field semantic checks
```go
// Source: https://pkg.go.dev/github.com/go-playground/validator/v10
validate := validator.New(validator.WithRequiredStructEnabled())

type Scope struct {
    InScope    []string `validate:"required,min=1,dive,min=5"`
    OutOfScope []string `validate:"required,min=1,dive,min=5"`
}

if err := validate.Struct(scope); err != nil {
    return err // map each field error to stable rule_id
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Regex/manual markdown scanning | CommonMark AST parsing with `goldmark` | Ongoing ecosystem standardization | Fewer false positives/negatives in required section detection. |
| Monolithic pass/fail validation | Structured reports with severity + machine-readable rules | Modern CLI validation workflows | Enables targeted guided fixes and policy-safe blocking. |
| Frontmatter parsed separately from markdown pipeline | Parser-integrated frontmatter extensions | Recent Go markdown tooling maturity | Cleaner parse pipeline and less parser glue code. |

**Deprecated/outdated:**
- Treating semantic checks as optional warnings-only quality hints for this workflow: incompatible with project fail-closed objective.

## Open Questions

1. **Boundary quality threshold details**
   - What we know: `in-scope` and `out-of-scope` sections must exist and be explicit.
   - What's unclear: exact minimum quality rubric (for example, min item count, prohibited vague phrases).
   - Recommendation: define and version a concrete rubric in this phase plan (`VAL.SCOPE.*` rules) and test with fixtures.

2. **Schema source of truth format**
   - What we know: strict schema validation is required.
   - What's unclear: whether canonical schema artifact should be JSON Schema only, CUE only, or both.
   - Recommendation: use JSON Schema as runtime gate in Phase 04; consider CUE only if later phases need schema composition ergonomics.

## Sources

### Primary (HIGH confidence)
- https://developers.openai.com/codex/skills - required skill metadata (`name`, `description`) and boundary guidance in descriptions.
- https://pkg.go.dev/github.com/yuin/goldmark - CommonMark-compliant parser and extension model.
- https://pkg.go.dev/go.abhg.dev/goldmark/frontmatter - parser-integrated frontmatter decoding patterns.
- https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6 - strict schema validation capabilities and error model.
- https://pkg.go.dev/github.com/go-playground/validator/v10 - semantic/cross-field validation patterns.
- https://pkg.go.dev/charm.land/huh/v2 - interactive prompt flow primitives for targeted one-at-a-time follow-up.
- https://pkg.go.dev/gopkg.in/yaml.v3 - YAML parsing support metadata.

### Secondary (MEDIUM confidence)
- https://pkg.go.dev/github.com/adrg/frontmatter - alternative frontmatter parser used for tradeoff comparison.

### Tertiary (LOW confidence)
- Web ecosystem searches for "Go markdown/frontmatter/json-schema libraries 2026" used for discovery only; critical claims above were verified against package or official documentation.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - core libraries and versions are documented on official package pages and align with existing project stack research.
- Architecture: HIGH - patterns directly enforce locked decisions (fail-closed, one-at-a-time issues, single-edit revalidate loop).
- Pitfalls: MEDIUM-HIGH - grounded in common validator workflow failure modes; exact frequency depends on implementation quality.

**Research date:** 2026-03-10
**Valid until:** 2026-04-09
