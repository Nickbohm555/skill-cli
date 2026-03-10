# Architecture Research

**Domain:** Go CLI skill generator (single-user, local-first)
**Researched:** 2026-03-10
**Confidence:** MEDIUM-HIGH

## Standard Architecture

### System Overview

```
┌────────────────────────────────────────────────────────────────────────────┐
│                           Presentation Layer                               │
├────────────────────────────────────────────────────────────────────────────┤
│  Cobra CLI     Prompt UI      Approval Diff/Preview     Exit/Error Policy │
│  (commands)    (adaptive Q&A) (before install)          (fail closed)     │
├────────────────────────────────────────────────────────────────────────────┤
│                           Application Layer                                │
├────────────────────────────────────────────────────────────────────────────┤
│  Orchestrator Pipeline                                                    │
│  ingest-url -> extract-doc-model -> question-loop -> draft-skill          │
│      -> strict-validate -> overlap-check -> install-plan -> confirm       │
├────────────────────────────────────────────────────────────────────────────┤
│                             Domain Layer                                   │
├────────────────────────────────────────────────────────────────────────────┤
│  SkillSpec   ValidationRules   OverlapPolicy   InstallPolicy   RiskModel   │
├────────────────────────────────────────────────────────────────────────────┤
│                            Adapter Layer                                   │
├────────────────────────────────────────────────────────────────────────────┤
│ HTTP Fetch  HTML Extract  LLM/Generator  SkillStore FS  Atomic Installer  │
│ Existing Skill Indexer + Similarity Engine + Optional GitHub installer     │
└────────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| CLI command surface | Parse args/flags, route command, set non-zero exit codes | `cmd/` with Cobra `RunE` |
| Session orchestrator | Drive the full generation state machine and retries | `internal/app/pipeline` |
| URL ingestion | Download one URL with timeout, size/type checks, normalization | `net/http` client + guardrails |
| Content extractor | Convert raw HTML into compact structured context | goquery/readability adapter |
| Adaptive questioning | Ask targeted follow-ups based on missing/weak fields | prompt adapter + question policy engine |
| Skill draft generator | Produce candidate `SKILL.md` and optional folder skeleton | deterministic template + model adapter |
| Strict validator | Enforce required frontmatter, scope boundaries, quality gates | parser + rule engine (+ optional JSON Schema/CUE) |
| Overlap analyzer | Detect semantic/scope overlap with installed skills | indexed metadata + lexical/semantic similarity |
| Install planner | Build explicit plan: create/update/merge/reject | policy engine with risk scoring |
| Safe installer | Stage, preview, confirm, write atomically into skill home | temp dir + atomic rename + rollback hooks |
| Skill registry adapter | Resolve and enumerate skill roots (`$CODEX_HOME/skills`, repo/user scopes) | path resolver + filesystem abstraction |
| Audit/log subsystem | Record decisions and diagnostics for reproducibility | local JSONL logs in app data dir |

## Recommended Project Structure

```
cmd/
└── skill-weaver/
    └── main.go                 # Cobra entrypoint

internal/
├── cli/                        # command wiring, prompt rendering, output formatting
│   ├── command/
│   └── ui/
├── app/                        # orchestrated use-cases (pipelines)
│   ├── generate/
│   ├── validate/
│   ├── overlap/
│   └── install/
├── domain/                     # pure domain rules (no IO)
│   ├── skill/
│   ├── validation/
│   ├── overlap/
│   └── install/
├── adapters/                   # side-effect boundaries
│   ├── httpfetch/
│   ├── htmlextract/
│   ├── generator/
│   ├── skillstore/
│   ├── similarity/
│   └── atomicfs/
└── platform/                   # cross-cutting concerns
    ├── config/
    ├── logging/
    └── telemetry/

schemas/
└── skill.schema.json           # strict output contract (or CUE equivalent)

test/
├── fixtures/
│   ├── html/
│   ├── skills/
│   └── overlap-cases/
└── e2e/
    └── generate_install_test.go
```

### Structure Rationale

- **`internal/domain/` first:** keeps validation and overlap logic testable and deterministic.
- **`internal/app/` second:** expresses pipeline state transitions explicitly, including retry loops.
- **`internal/adapters/` isolation:** lets ingestion, extraction, model, and filesystem evolve without rewriting core logic.
- **`schemas/` as contract:** strict output gates become auditable and versioned.

## Architectural Patterns

### Pattern 1: Pipeline + State Machine

**What:** A linear pipeline with explicit stage outcomes (`pass`, `retry`, `abort`) and bounded retries.
**When to use:** Interactive CLIs with hard quality gates and multi-step refinement.
**Trade-offs:** Very clear control flow, but requires careful state typing to avoid spaghetti transitions.

**Example:**
```go
type StageResult int

const (
    Pass StageResult = iota
    Retry
    Abort
)

func Run(session *Session) error {
    if err := Ingest(session); err != nil { return err }
    for {
        AskAdaptiveQuestions(session)
        DraftSkill(session)
        result, err := ValidateStrict(session)
        if err != nil { return err }
        if result == Pass { break }
        if result == Abort { return ErrValidationAborted }
    }
    return PlanAndInstall(session)
}
```

### Pattern 2: Ports and Adapters (Hexagonal-lite)

**What:** Domain/app layers depend on interfaces; adapters implement external IO.
**When to use:** Greenfield tools needing strong testability and safe filesystem operations.
**Trade-offs:** More interface boilerplate, but much safer refactors and easier fake adapters in tests.

**Example:**
```go
type SkillStore interface {
    List(ctx context.Context) ([]SkillRef, error)
    Stage(ctx context.Context, candidate SkillArtifact) (StageRef, error)
    InstallAtomic(ctx context.Context, stage StageRef, target string) error
}
```

### Pattern 3: Validate-Then-Commit

**What:** Never write to final skill location before all validators and overlap policy checks pass.
**When to use:** Any flow that mutates user-managed local registries.
**Trade-offs:** Slightly slower due to staging, but prevents corrupt or partial installs.

## Data Flow

### Request Flow

```
User runs `skill-weaver generate --url <doc-url>`
    ↓
CLI parses command + initializes session context
    ↓
URL Ingestion fetches page (timeouts/size/content-type checks)
    ↓
Extractor builds normalized document model (title, sections, examples, commands)
    ↓
Question Engine asks adaptive follow-ups for missing/weak skill fields
    ↓
Draft Generator produces candidate skill artifact
    ↓
Strict Validator checks:
  - required frontmatter (name, description)
  - bounded capability scope
  - instruction quality gates
  - schema/policy conformance
    ↓
Overlap Analyzer compares candidate against installed skills index
    ↓
Install Planner proposes action (new / merge-update / reject)
    ↓
User approval gate (show diff + risks)
    ↓
Safe Installer stages artifact and atomically installs into resolved skill root
    ↓
Post-install verification (can be discovered + parses correctly) and success output
```

### State Management

```
SessionState
  ├─ Input: URL, user answers
  ├─ Derived: doc model, candidate skill, validation findings
  ├─ External: existing skills index, install target
  └─ Decision: plan + approval + final install status

State transitions are append-only in memory and mirrored to local JSONL logs for replay/debug.
```

### Key Data Flows

1. **Ingestion-to-question flow:** URL content becomes structured evidence used to generate only the minimum necessary questions.
2. **Validation feedback loop:** Validator findings are transformed into precise follow-up questions until the artifact passes or user aborts.
3. **Overlap-to-install flow:** Similarity findings produce explicit policy decisions before any filesystem mutation.

## Build Order and Dependencies

1. **Domain contracts first (blocking dependency)**
   - `SkillSpec`, `ValidationReport`, `OverlapReport`, `InstallPlan`.
   - Everything else depends on stable domain types.

2. **Strict validator second**
   - Build hard fail-fast quality gates before generation ergonomics.
   - Prevents "demo-first, rewrite-later" architecture debt.

3. **Ingestion + extraction third**
   - One-URL fetch and normalized doc model powering downstream stages.
   - Depends on domain contracts and validation target shape.

4. **Adaptive questioning and draft generation fourth**
   - Uses extractor outputs and validator feedback loop.
   - Should be implemented only after validation semantics are fixed.

5. **Overlap analyzer fifth**
   - Depends on candidate artifact and skill index adapter.
   - Produces `InstallPlan` inputs.

6. **Safe installer sixth**
   - Final mutating stage; depends on overlap policy and approval UX.
   - Must include staging + atomic commit semantics from day one.

7. **End-to-end workflow + fixtures last**
   - Proves full flow (URL -> approved installed skill) under realistic failure cases.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1 user / local | Single process monolith, local index, file-based logs |
| tens of users (same machine pattern) | Keep monolith; add config profiles and stronger cache/indexing |
| multi-user/service future | Split installer/indexer behind API boundary; keep domain logic shared |

### Scaling Priorities

1. **First bottleneck:** overlap checks become slow as skill count grows -> add persisted lightweight index and incremental refresh.
2. **Second bottleneck:** adaptive loop latency from generation/validation churn -> cache extraction and rule-evaluation artifacts per session.

## Anti-Patterns

### Anti-Pattern 1: Write-before-validate

**What people do:** Write generated skill directly into final skill directory, then try to validate/fix in place.
**Why it's wrong:** Leaves partial or invalid skills installed; hard to recover trust and state.
**Do this instead:** Stage externally, validate completely, then atomically install only approved artifacts.

### Anti-Pattern 2: Coupling prompts to business rules

**What people do:** Encode validation and overlap policy inside prompt text branching.
**Why it's wrong:** Rules become untestable and drift from expected policy.
**Do this instead:** Keep prompt engine thin; all policy decisions live in deterministic domain validators.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Target docs URL | HTTP GET with bounded retries, timeout, max body size | Enforce allowlist/denylist later if needed |
| Optional model backend | Adapter interface (`Draft(ctx, prompt, constraints)`) | Keep replaceable to avoid lock-in |
| OpenAI skill ecosystem docs | Read-only references for schema expectations | Used for policy updates, not runtime dependency |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `cli` ↔ `app` | function calls with typed request/response | No domain rules in CLI package |
| `app` ↔ `domain` | pure structs + interfaces | Domain remains side-effect free |
| `app` ↔ `adapters` | ports/interfaces | Enables in-memory test doubles |
| `install` ↔ `skillstore` | staged artifact protocol | Enforces approval and atomicity |

## Sources

- [Cobra command architecture and `RunE` patterns](https://cobra.dev/docs/how-to-guides/working-with-commands) (official docs, HIGH)
- [Cobra package reference and feature set](https://pkg.go.dev/github.com/spf13/cobra) (official package docs, HIGH)
- [Codex skills format and discovery behavior](https://developers.openai.com/codex/skills) (official docs, HIGH)
- [Codex customization paths (`~/.codex/AGENTS.md`, `$HOME/.agents/skills`)](https://developers.openai.com/codex/concepts/customization) (official docs, HIGH)
- [Skill installer behavior for `$CODEX_HOME/skills` defaulting to `~/.codex/skills`](file:///Users/nickbohm/.codex/skills/.system/skill-installer/SKILL.md) (local system skill spec, MEDIUM-HIGH)
- [Go `os.Rename` caveats (cross-dir/platform atomicity constraints)](https://pkg.go.dev/os#Rename) (official Go docs, HIGH)
- [Go `os.CreateTemp` behavior for safe staging](https://pkg.go.dev/os#CreateTemp) (official Go docs, HIGH)
- [goquery parser capabilities and UTF-8 caveat](https://pkg.go.dev/github.com/PuerkitoBio/goquery) (official package docs, HIGH)
- [CUE validation and unification model](https://cuelang.org/docs/concept/how-cue-enables-data-validation/) (official docs, HIGH)

---
*Architecture research for: CLI skill generator domain (Skill Weaver)*
*Researched: 2026-03-10*
