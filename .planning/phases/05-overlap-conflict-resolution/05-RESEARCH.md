# Phase 05: Overlap & Conflict Resolution - Research

**Researched:** 2026-03-10
**Domain:** Overlap detection and explicit conflict-resolution UX for generated Codex skills
**Confidence:** HIGH

## Summary

Phase 05 should be planned as a deterministic decision stage between Phase 04 validation and Phase 06 install approval. The core outcome is not file mutation; it is a resolved conflict decision artifact that clearly states whether the generated skill is a `new_install`, `update_existing`, `merge_with_existing`, or `abort`.

The practical implementation path is to compare a normalized candidate skill against an index of installed skills using layered heuristics: hard collisions first (exact name/path), then structural overlap (frontmatter + section similarity), then scope conflict signals (in-scope and out-of-scope intersections). Only explicit user choice can resolve overlap states; unresolved findings must fail closed and block progression.

This phase should output an auditable resolution summary shown to the user before any write/install action in Phase 06. The biggest planning risk is ambiguity in merge/update semantics, so tasks should define resolution modes, confidence/risk scoring, and exact CLI prompts up front.

**Primary recommendation:** Build a two-step conflict engine (`detect -> decide`) that emits a machine-readable `ConflictResolutionDecision` consumed by Phase 06, with mandatory explicit user selection whenever overlap severity is not `none`.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (`os`, `path/filepath`) | Go stdlib | Enumerate installed skills and resolve deterministic identifiers | Required for local-first registry scanning and path-safe comparisons. |
| `github.com/yuin/goldmark` | `v1.7.16` | Parse `SKILL.md` into normalized sections for overlap comparison | Reuses Phase 04 parser stack to avoid divergent parsing logic. |
| `go.abhg.dev/goldmark/frontmatter` | `v0.3.0` | Extract `name` and `description` metadata used for collision checks | Provides typed frontmatter extraction used in both validation and overlap analysis. |
| `charm.land/huh/v2` | `v2.0.3` | Explicit CLI decision prompts (`Select`, `Confirm`) | Supports structured, explicit user choices for merge/update/abort flows. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/google/go-cmp` | `v0.7.0` | Compare normalized skill structs for exact-equivalence checks | Use to quickly classify "candidate equals installed skill" as low-risk update/no-op path. |
| `github.com/sergi/go-diff` | `v1.4.0` | Generate compact textual delta previews for decision context | Use when showing "what would change" between installed and candidate instructions. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Heuristic lexical overlap scoring | Embedding/vector similarity service | Better semantic recall, but adds non-local dependency and unpredictability for v1 local-first CLI. |
| `huh` interactive forms | Raw stdin prompts (`fmt.Scanln`) | Lower dependency surface, but weaker multi-option UX and higher input-validation complexity. |
| `go-diff` text differ | Custom line-by-line differ | Fewer dependencies, but easier to produce noisy or misleading change previews. |

**Installation:**
```bash
go get github.com/yuin/goldmark@v1.7.16
go get go.abhg.dev/goldmark/frontmatter@v0.3.0
go get charm.land/huh/v2@v2.0.3
go get github.com/google/go-cmp/cmp@latest
go get github.com/sergi/go-diff@v1.4.0
```

## Architecture Patterns

### Recommended Project Structure
```text
internal/overlap/
├── model.go                 # overlap entities, severity levels, decision contract
├── index_installed.go       # scan + parse installed skills into normalized index
├── detect.go                # overlap detection pipeline and scoring
├── classify.go              # map raw signals into severity/risk categories
├── decision_flow.go         # explicit user-choice orchestration (no writes)
└── resolution_summary.go    # final "selected outcome" output for Phase 06 handoff

internal/app/generate/
└── overlap_stage.go         # orchestrator stage: validate pass -> overlap -> resolution artifact
```

### Pattern 1: Layered Overlap Detection
**What:** Run checks from highest certainty to lowest: exact collision -> structural overlap -> semantic scope conflict.
**When to use:** For every validated candidate entering Phase 05.
**Example:**
```go
type OverlapSeverity string

const (
    SeverityNone   OverlapSeverity = "none"
    SeverityLow    OverlapSeverity = "low"
    SeverityMedium OverlapSeverity = "medium"
    SeverityHigh   OverlapSeverity = "high"
)

func Detect(candidate SkillProfile, installed []SkillProfile) []OverlapFinding {
    findings := []OverlapFinding{}
    for _, existing := range installed {
        if candidate.Name == existing.Name {
            findings = append(findings, NewFinding("OVLP.NAME.EXACT", SeverityHigh, existing.ID))
            continue
        }
        score := SimilarityScore(candidate, existing)
        if score >= 0.70 {
            findings = append(findings, NewFinding("OVLP.SCOPE.SEMANTIC", SeverityMedium, existing.ID))
        }
    }
    return findings
}
```

### Pattern 2: Decision Contract as a Boundary Artifact
**What:** Convert overlap findings + user choice into a typed decision object consumed by later phases.
**When to use:** Immediately after explicit user resolution choice.
**Example:**
```go
type ResolutionMode string

const (
    ResolutionNewInstall ResolutionMode = "new_install"
    ResolutionUpdate     ResolutionMode = "update_existing"
    ResolutionMerge      ResolutionMode = "merge_with_existing"
    ResolutionAbort      ResolutionMode = "abort"
)

type ConflictResolutionDecision struct {
    CandidateID      string
    TargetSkillID    string
    Mode             ResolutionMode
    SelectedAt       time.Time
    RequiresDiffView bool
    Blocking         bool // true until user made an explicit choice
    Notes            string
}
```

### Pattern 3: Explicit Choice UX with Outcome Preview
**What:** Force user selection from explicit options and echo the exact selected outcome before progressing.
**When to use:** Any time overlap severity is not `none`.
**Example:**
```go
var mode ResolutionMode
err := huh.NewSelect[ResolutionMode]().
    Title("Overlap detected. Choose a conflict-resolution path before install:").
    Options(
        huh.NewOption("Update existing skill", ResolutionUpdate),
        huh.NewOption("Merge candidate + existing", ResolutionMerge),
        huh.NewOption("Abort (no install)", ResolutionAbort),
    ).
    Value(&mode).
    Run()
if err != nil {
    return err
}
fmt.Printf("Selected outcome: %s\n", mode) // mandatory OVLP-03 surface before Phase 06
```

### Design Options and Tradeoffs

| Option | Strengths | Weaknesses | Recommendation |
|--------|-----------|------------|----------------|
| Name-only collision checks | Fast, deterministic, simple to explain | Misses semantic overlaps with different names | Use only as first-pass hard signal, not full detector. |
| Weighted lexical/structural heuristics | Local-first, explainable scoring, testable fixtures | Needs threshold tuning and can miss deep semantic overlap | **Recommended for v1** with conservative thresholds + explicit user confirmation. |
| Embeddings-based similarity | Better semantic recall on paraphrased skills | External service or model dependency, less deterministic, higher complexity | Defer to v1.x after baseline overlap UX is stable. |

### Anti-Patterns to Avoid
- **Auto-resolve on high overlap:** violates explicit choice requirement and can mutate trusted skills unexpectedly.
- **Opaque confidence score with no signal breakdown:** users cannot make informed merge/update decisions.
- **Mixing overlap decision and file writes in same stage:** breaks phase boundary and makes fail-closed enforcement brittle.
- **Silent fallthrough to install when prompts are interrupted:** cancellations must map to `abort` and remain blocked.

## Suggested Data Structures and Comparison Heuristics

### Core data model

```go
type SkillProfile struct {
    ID          string   // stable ID from path/name
    Name        string
    Description string
    InScope     []string
    OutOfScope  []string
    Commands    []string
    SourcePath  string
}

type OverlapFinding struct {
    RuleID      string
    ExistingID  string
    Severity    OverlapSeverity
    Score       float64
    Signals     []string // e.g. "name_exact", "in_scope_jaccard:0.82"
    Explanation string
}
```

### Heuristic scoring (practical v1 formula)

Use weighted signals in a bounded score:

- `name_exact` (hard override): severity `high`
- `description_similarity` (normalized token overlap): weight `0.35`
- `in_scope_jaccard`: weight `0.30`
- `out_of_scope_conflict` (candidate in-scope intersects existing out-of-scope or inverse): weight `0.20`
- `command_overlap`: weight `0.15`

Classification recommendation:

- `score < 0.40` -> `none/low` (allow `new_install` default path with explicit confirmation)
- `0.40 <= score < 0.70` -> `medium` (require explicit update/merge/abort choice)
- `score >= 0.70` or hard collision -> `high` (strong warning; no implicit path)

### Why these heuristics work for planning

- Deterministic and local (no remote model needed).
- Explainable to user with concrete signals.
- Easy to test with fixtures and tune thresholds without architecture rewrites.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Interactive multi-option conflict prompts | Ad hoc stdin loops with manual parsing | `huh` `Select`/`Confirm` fields | Better validation, consistent UX, and fewer input edge-case bugs. |
| Markdown/frontmatter parsing split across phases | Separate ad hoc parser for overlap stage | Reuse Phase 04 parser contract (`goldmark` + frontmatter) | Avoids parse drift between validation and overlap decisions. |
| Human-readable textual change previews | Custom diff formatting logic | `go-diff` | Mature differ reduces misleading preview output and edge-case handling. |

**Key insight:** The risk in this phase is incorrect decisions, not raw algorithm speed; use well-known parsing/prompt/diff primitives and keep policy logic deterministic.

## CLI UX Patterns for Explicit User Choice

### Required interaction contract

1. Show overlap findings with severity and plain-language explanation.
2. If severity `none`, still show "no conflicts found" and proceed to Phase 06 pre-approval.
3. If overlap exists, present explicit choices (`update`, `merge`, `abort`) as mutually exclusive options.
4. Echo selected outcome in a dedicated "Resolution Summary" block before any write/install action.
5. Require reconfirmation if user changes choice after viewing diff/summary.

### Recommended prompt flow

```text
[Overlap Findings]
- Existing skill: shell-helper
- Severity: HIGH
- Signals: name_exact, in_scope_jaccard=0.78

Choose resolution path:
  1) Update existing skill
  2) Merge with existing skill
  3) Abort

[Resolution Summary]
Selected: Merge with existing skill
Target: shell-helper
Next step: Show merged-preview in Phase 06 approval gate
Status: BLOCKED until user confirms continue
```

### UX tradeoff notes

- Defaulting to `abort` on interruption is safer than defaulting to `update`.
- Showing one target skill at a time simplifies decisions when multiple overlaps exist.
- "Why this was flagged" text is mandatory for user trust and debuggability.

## Integration Boundaries

### Input contract from Phase 4 (Validation & Quality Gates)

Phase 05 should only run when:
- Candidate passed all blocking validation errors.
- Candidate profile is normalized and parse-stable (same structure used in overlap stage).
- Validation report is attached for traceability but not re-adjudicated here.

If Phase 04 still has unresolved blocking errors, Phase 05 must not start (fail closed).

### Output contract to Phase 6 (Approval-Gated Install & Activation)

Phase 05 must emit:
- `OverlapReport` (all findings + signals + severity)
- `ConflictResolutionDecision` (explicit selected mode)
- `ResolutionSummary` (human-readable block for pre-install display)

Phase 06 is responsible for:
- Rendering final file diff/preview for install/write operations
- Approval confirmation and filesystem mutation
- Enforcing install block when Phase 05 decision is absent or `abort`

Phase 05 must not:
- Write to `$CODEX_HOME/skills`
- Modify existing installed skills
- Auto-commit merge/update results

## Common Pitfalls

### Pitfall 1: False positives from generic language
**What goes wrong:** Generic words ("best practices", "guide") inflate similarity.
**Why it happens:** Overweighting description token overlap.
**How to avoid:** Downweight stopwords and require at least one strong scope/command signal.
**Warning signs:** Many medium/high flags for clearly unrelated skills.

### Pitfall 2: False negatives from naming diversity
**What goes wrong:** Conflicting skills with different names bypass checks.
**Why it happens:** Name collision treated as primary and only detector.
**How to avoid:** Include in-scope/out-of-scope and command overlap signals.
**Warning signs:** Duplicate behavior slips through when names differ.

### Pitfall 3: Ambiguous merge semantics
**What goes wrong:** "Merge" means different operations to users and code.
**Why it happens:** No explicit merge policy definition in decision contract.
**How to avoid:** Define merge policy modes (e.g., append sections, replace matching headings) before implementation tasks.
**Warning signs:** Planner tasks mention merge without deterministic transformation rules.

### Pitfall 4: Decision state lost between stages
**What goes wrong:** Phase 06 cannot prove explicit user resolution happened.
**Why it happens:** Decision stored only in transient UI state.
**How to avoid:** Persist decision artifact in session state with timestamp and target IDs.
**Warning signs:** Install flow re-prompts overlap or proceeds without a stored decision.

## Edge Cases and Testing Strategy

### Critical edge cases

1. Candidate exactly equals installed skill content (`no-op` style overlap).
2. Candidate has unique name but near-identical scope to one installed skill.
3. One candidate overlaps multiple installed skills with different severities.
4. User cancels prompt (Ctrl+C/EOF) during choice flow.
5. Installed skill parse failure (malformed `SKILL.md`) while building index.
6. Case-sensitivity/path-normalization differences on skill names across filesystems.
7. Candidate marked "update" but selected target missing by time of Phase 06.

### Test matrix (planning-ready)

| Test Type | What to Verify | Example |
|----------|-----------------|---------|
| Unit: scoring | Deterministic score + severity classification | same fixture always yields same `OVLP.*` findings |
| Unit: decision policy | Block when overlap unresolved | missing user choice sets `Blocking=true` |
| Unit: edge behavior | interruption maps to safe abort | prompt cancel -> `ResolutionAbort` |
| Integration: phase handoff | Phase 04 pass required before detection | unresolved `Error` from validation skips Phase 05 |
| Integration: output contract | Phase 06 receives complete decision artifact | install stage rejects absent `ConflictResolutionDecision` |
| Golden fixtures | known overlap scenarios remain stable | `fixtures/overlap-cases/*.yaml` expected report snapshots |
| UX snapshot tests | prompt text includes explicit choices and summary | selection screen + summary output unchanged unless intentional |

### Verification commands (once implementation exists)

```bash
go test ./internal/overlap -v
go test ./internal/app/generate -run Overlap -v
go test ./... -run "Overlap|Conflict|InstallGate" -v
```

## Code Examples

Verified patterns from official sources:

### Parse frontmatter metadata for profile extraction
```go
// Source: https://pkg.go.dev/go.abhg.dev/goldmark/frontmatter
md := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
ctx := parser.NewContext()
_ = md.Convert(src, io.Discard, parser.WithContext(ctx))
fm := frontmatter.Get(ctx)
```

### Explicit CLI conflict choice
```go
// Source: https://pkg.go.dev/charm.land/huh/v2
var choice string
_ = huh.NewSelect[string]().
    Title("Choose conflict resolution:").
    Options(
        huh.NewOption("Update", "update"),
        huh.NewOption("Merge", "merge"),
        huh.NewOption("Abort", "abort"),
    ).
    Value(&choice).
    Run()
```

### Struct equality checks for exact-overlap classification
```go
// Source: https://pkg.go.dev/github.com/google/go-cmp
if cmp.Equal(candidateProfile, existingProfile) {
    findings = append(findings, NewFinding("OVLP.EXACT.CONTENT", SeverityHigh, existingProfile.ID))
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Name-only duplicate detection | Multi-signal overlap analysis + explicit decision UX | Modern local-agent CLI workflows | Reduces silent conflicts and improves user trust in install decisions. |
| Implicit overwrite/update behavior | Approval-gated explicit resolution and outcome preview | Current Codex/agent safety patterns | Aligns with fail-closed local-first safety constraints. |
| Unstructured prompt text for risky actions | Structured select/confirm interaction patterns | Current CLI UX tooling maturity | Makes choices testable and deterministic. |

**Deprecated/outdated:**
- Silent overwrite as default conflict behavior: incompatible with OVLP-02 and fail-closed policy.

## Open Questions

1. **Merge semantics at section granularity**
   - What we know: Phase 05 must offer merge as an explicit path.
   - What's unclear: exact deterministic merge algorithm for `SKILL.md` sections.
   - Recommendation: In planning, define merge policy now as decision metadata only; actual materialized merge transform can be finalized in Phase 6 where preview/install pipeline exists.

2. **Threshold tuning source**
   - What we know: heuristic thresholds are required for practical overlap severity classes.
   - What's unclear: initial threshold values that minimize false positives for real user skill sets.
   - Recommendation: start with conservative defaults from this doc and include fixture-driven calibration task before locking values.

## Sources

### Primary (HIGH confidence)
- https://developers.openai.com/codex/skills - skill format expectations and scope-boundary importance.
- https://pkg.go.dev/charm.land/huh/v2 - explicit terminal choice/confirm primitives.
- https://pkg.go.dev/github.com/google/go-cmp - deterministic deep comparison support for profile equivalence checks.
- https://pkg.go.dev/github.com/sergi/go-diff - text diff generation primitives for change previews.
- https://pkg.go.dev/github.com/yuin/goldmark - markdown parser for normalized skill section extraction.
- https://pkg.go.dev/go.abhg.dev/goldmark/frontmatter - frontmatter extraction for name/description profile fields.

### Secondary (MEDIUM confidence)
- https://cobra.dev/docs/how-to-guides/working-with-commands/ - command/error handling patterns informing explicit CLI stage flow.

### Tertiary (LOW confidence)
- None. Critical claims are tied to official documentation and local project constraints.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - selected packages are official docs-backed and align with prior project architecture research.
- Architecture: HIGH - patterns directly map to OVLP-01/02/03 and phase boundaries.
- Pitfalls: MEDIUM-HIGH - failure modes are common in conflict pipelines and covered by deterministic tests.

**Research date:** 2026-03-10
**Valid until:** 2026-04-09
