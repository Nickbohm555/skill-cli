# Project Research Summary

**Project:** Skill Weaver
**Domain:** Local-first Go CLI that generates installable Codex skills from a single docs URL
**Researched:** 2026-03-10
**Confidence:** MEDIUM-HIGH

## Executive Summary

Skill Weaver is a quality-gated developer tool, not a generic content converter. The research converges on a deterministic pipeline: ingest one URL, extract high-signal content, run adaptive Q&A to close gaps, generate a candidate skill, enforce strict validation, check overlap against installed skills, and only then allow approval-gated install. Experts build this kind of product as a fail-closed workflow with explicit policies at each transition, because local registry mutation is the highest-risk operation.

The recommended implementation approach is a Go monolith with clear layer boundaries (CLI, app pipeline, domain policy, adapters), using Cobra for commands, Huh for interactive flows, and strict schema plus semantic validators as non-negotiable gates. The strongest delivery strategy is to lock domain contracts and validation semantics first, then add ingestion/extraction, then adaptive generation, then conflict handling and transactional install, followed by adversarial hardening.

The key risks are untrusted remote content, incomplete extraction from dynamic docs pages, weak semantic validation, and non-atomic install writes. Mitigation is consistent across the research set: treat fetched content as untrusted input, add extraction completeness diagnostics, enforce semantic quality rubrics with targeted re-ask loops, resolve overlap pre-install, and stage/atomically commit filesystem changes with rollback paths.

## Key Findings

### Recommended Stack

The stack is mature and intentionally conservative for reliability: Go 1.25.x, Cobra + Viper for CLI/config, Huh v2 for adaptive prompts, and `openai-go/v3` for generation. Supporting choices focus on hardening and testability: goquery for deterministic single-page extraction, retryable HTTP for network resilience, schema + semantic validators for output quality, and `afero` for safe filesystem testing.

Critical version notes: Go 1.25.x satisfies newer package floors (for example, goquery and validator requirements), and Charm libraries should stay aligned on v2 majors to avoid UI integration friction.

**Core technologies:**
- Go 1.25.x: runtime and distribution base - stable CLI ecosystem with strong stdlib for IO/process safety.
- `github.com/spf13/cobra`: command UX and routing - de-facto standard with predictable help/completion ergonomics.
- `charm.land/huh/v2`: adaptive interactive flow - strong fit for iterative ask/validate/re-ask loops.
- `github.com/openai/openai-go/v3`: draft generation adapter - official SDK for constrained generation retries.
- `github.com/spf13/viper`: config/env handling - standard with Cobra for runtime and key management.

### Expected Features

v1 table stakes are consistent: deterministic single-URL ingestion, robust extraction/normalization, adaptive refinement tied to strict validation, overlap detection with explicit resolution, and approval-gated install with preview/diff. Differentiators worth building after core stability are adaptive follow-ups from specific validator failures, overlap-aware merge planning, and atomicity scoring to keep generated skills narrow and composable.

Research strongly advises deferring multi-page crawling, team governance workflows, and default one-shot generation in early releases because they broaden ambiguity and degrade reliability before core quality loops are proven.

**Must have (table stakes):**
- Single URL ingest + deterministic extraction - base for all downstream quality.
- Adaptive Q&A linked to strict validation - closes missing/weak fields before output.
- Overlap detection with explicit merge/update/rename/cancel decision path - prevents registry confusion.
- Approval-gated install with preview/diff - preserves user trust and control.
- Fail-closed behavior on uncertainty - no install when validation or conflict confidence is low.

**Should have (competitive):**
- Validation-failure-driven adaptive questioning - highest leverage correctness boost.
- Merge/update planner with impact preview - safer evolution of existing skills.
- Session persistence for resumable refinement - better UX for iterative quality improvements.
- Install simulation with rollback helper - lower risk for update flows.

**Defer (v2+):**
- Multi-page/domain crawling.
- Team/shared governance workflows.
- IDE companion experiences.

### Architecture Approach

The architecture should be pipeline-first and policy-first: use a typed state machine in `internal/app`, keep all validation/overlap/install rules in deterministic `internal/domain`, and isolate side effects behind adapters (fetch/extract/generate/store/install). The strongest pattern across sources is validate-then-commit: never mutate final skill paths before all quality and policy gates pass.

**Major components:**
1. CLI/prompt layer - command entry, adaptive prompts, approval UX, and explicit exit semantics.
2. Orchestrator pipeline - stage sequencing with bounded retry/abort transitions.
3. Domain policy engine - skill spec rules, semantic validation, overlap and install policy scoring.
4. Ingestion/extraction adapters - URL fetch guardrails and deterministic content normalization.
5. Registry/install subsystem - staged writes, atomic move, rollback, and post-install verification.

### Critical Pitfalls

The highest-probability failures are known and preventable if handled early in phase order.

1. **Untrusted docs content influences behavior** - treat fetched content as untrusted data, sanitize aggressively, and test prompt-injection fixtures.
2. **Incomplete extraction from dynamic docs** - add completeness heuristics and require explicit user confirmation on low-signal pages.
3. **"Valid-looking" but semantically weak skills** - enforce semantic rubrics (scope boundaries, non-goals, trigger clarity) and targeted re-asks.
4. **Late overlap/collision handling** - run overlap policy before approval and require explicit resolution choices.
5. **Non-atomic install writes** - use staging + atomic move + rollback; fail closed on any filesystem error.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Secure Ingestion and Domain Contracts
**Rationale:** Everything downstream depends on safe URL handling and stable domain types; this minimizes rework and security debt.
**Delivers:** `SkillSpec`/validation/overlap/install contract types, URL safety policy (scheme/host/redirect/timeout/size limits), baseline fetch adapter.
**Addresses:** Single-URL deterministic ingestion, fail-closed policy.
**Avoids:** SSRF/hostile redirect risk and untrusted-input ambiguity.

### Phase 2: Extraction Quality Gates
**Rationale:** Adaptive generation quality is capped by extraction quality; build diagnostics before prompting.
**Delivers:** HTML-to-doc model extraction, completeness heuristics, low-signal warnings, extraction preview diagnostics.
**Addresses:** Clean content extraction table stake.
**Avoids:** Building skills from nav/noise-heavy or incomplete pages.

### Phase 3: Semantic Validation and Adaptive Refinement
**Rationale:** Validation semantics must drive adaptive Q&A; this is the core product loop.
**Delivers:** strict schema + semantic validator, targeted re-ask engine, bounded retry/abort state machine.
**Addresses:** Interactive refinement, strict quality gate, narrow single-capability output.
**Avoids:** "Looks valid but unusable" skill artifacts.

### Phase 4: Conflict Resolution and Approval Contract
**Rationale:** Overlap/approval semantics are trust-critical and should precede any write path.
**Delivers:** existing-skill indexing, overlap scoring, merge/update/rename planner, interactive + non-interactive approval policy.
**Addresses:** collision detection, explicit user control, script-safe behavior.
**Avoids:** silent conflict drift and CI/TTY prompt mismatch failures.

### Phase 5: Transactional Install and Verification
**Rationale:** Write durability is a hard requirement for local-first tooling.
**Delivers:** staged artifact writes, atomic install, rollback hooks, post-install parse/discovery verification.
**Addresses:** safe install to `$CODEX_HOME/skills`.
**Avoids:** partial/corrupt installs and inconsistent registry state.

### Phase 6: Hardening, E2E Reliability, and v1.x Differentiators
**Rationale:** Once end-to-end flow is stable, add stress/security tests and selected P2 differentiators.
**Delivers:** adversarial fixtures, crash-interruption tests, performance caching, optional session resume and install simulation.
**Addresses:** resilience and operational confidence before broader scope.
**Avoids:** regressions from real-world noisy docs and repeated retry churn.

### Phase Ordering Rationale

- Policy and safety constraints come first because every later feature consumes those contracts.
- Extraction precedes adaptive generation to avoid optimizing prompts around bad source evidence.
- Overlap/approval is completed before install so all mutation decisions are explicit and auditable.
- Transactional install is isolated as its own phase to guarantee durability before expansion features.
- Hardening is intentionally last in build order but mandatory before release confidence.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 4:** overlap scoring thresholds and merge semantics may need deeper algorithm/provenance trade-off analysis.
- **Phase 6:** adversarial prompt-injection test design and red-team fixture coverage can benefit from focused security research.

Phases with standard patterns (skip research-phase):
- **Phase 1:** Go CLI command/config structure and network guardrails are well-documented.
- **Phase 2:** static HTML extraction plus diagnostics has established implementation patterns.
- **Phase 5:** staged-write + atomic rename + rollback is a mature filesystem pattern.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Versions and compatibility are well-sourced with official docs/releases. |
| Features | MEDIUM-HIGH | Strong cross-tool pattern evidence; some differentiators are inference-heavy for this niche workflow. |
| Architecture | MEDIUM-HIGH | Patterns are robust and practical, but final boundaries depend on implementation trade-offs. |
| Pitfalls | HIGH | Security and reliability risks are concrete, repeatedly documented, and directly testable. |

**Overall confidence:** MEDIUM-HIGH

### Gaps to Address

- Overlap scoring calibration: initial lexical/semantic thresholds need empirical tuning on real local skill corpora.
- Extraction coverage limits: JS-rendered/gated docs behavior needs explicit user-facing fallback policy validation.
- Non-interactive approval semantics: CI/automation force-flag behavior must be pinned with acceptance tests early.
- Optional deterministic/offline generation path: architecture supports it, but product priority and quality bar are unvalidated.

## Sources

### Primary (HIGH confidence)
- Go official release and stdlib docs (`go.dev`, `pkg.go.dev`) - runtime compatibility, filesystem and HTTP behavior.
- Official package documentation/releases for Cobra, Viper, Huh, OpenAI Go SDK, goquery, validator, jsonschema, afero.
- OpenAI Codex skills/customization docs - skill format and install/discovery expectations.
- OWASP LLM prompt-injection guidance - threat model and defensive patterns.

### Secondary (MEDIUM confidence)
- Claude Code and Aider docs - comparative UX/guardrail patterns for approval and iterative workflows.
- Firecrawl and Jina Reader docs - extraction behavior references relevant to URL ingestion strategy.
- Real-world CLI prompt-safety evidence (Netlify CLI PR) - practical confirmation pattern.

### Tertiary (LOW confidence)
- Internal/local skill specs and inferred design heuristics for this specific product domain - useful but requires implementation validation.

---
*Research completed: 2026-03-10*
*Ready for roadmap: yes*
