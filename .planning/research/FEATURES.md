# Feature Research

**Domain:** CLI tool that converts a documentation URL into an installable Codex skill
**Researched:** 2026-03-10
**Confidence:** MEDIUM-HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Deterministic single-URL ingestion | Users expect "paste URL -> get output", not manual copy/paste pipelines | LOW | Must accept one URL, fetch content, and return stable extraction results for repeat runs. |
| Clean content extraction to LLM-friendly text | Raw HTML/noisy nav content causes poor generated skills | MEDIUM | Main-content extraction and markdown normalization are baseline for usable prompt synthesis. |
| Interactive refinement loop | One-shot generation often misses critical constraints and install context | MEDIUM | Guided Q&A should fill missing required fields before generation is finalized. |
| Strict schema + quality validation gate | Users expect output that works immediately, not "mostly right" skill files | MEDIUM | Block output when required sections are missing/weak; require fix-up prompts before continue. |
| Safe write/install flow with explicit confirmation | CLI coding tools now default to explicit permission controls for risky actions | LOW | Require approval before write to `$CODEX_HOME/skills`; default to no silent install/overwrite. |
| Dry-run preview and diff before apply | Users expect visibility before filesystem changes in local agent tooling | LOW | Show generated `SKILL.md` preview and planned file operations before install. |
| Existing-skill collision detection | Skill registries become noisy quickly without duplicate/conflict checks | MEDIUM | Check by name and semantic overlap; surface clear conflict reason before install path step. |
| Resumable/iterative session UX | CLI users expect resume/continue workflows in modern coding assistants | MEDIUM | Persist session state so user can refine skill later instead of restarting from scratch. |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Adaptive questioning based on validation failures | Minimizes user effort while improving output correctness | HIGH | Dynamic follow-ups targeted to failed constraints outperform static questionnaires. |
| Overlap-aware merge/update planner | Prevents skill sprawl while preserving existing customizations | HIGH | Generate "new vs merge vs update" options with conflict explanation and impact preview. |
| Skill atomicity scorer (single-capability guardrail) | Produces higher-quality skills that are composable and reusable | MEDIUM | Detect over-broad scope and ask user to split before install. |
| Install simulation with rollback plan | Reduces trust barrier for local registry writes | MEDIUM | Simulate write path, report changed files, and produce rollback command before approval. |
| Provenance block in generated skill metadata | Makes long-term maintenance easier ("where did this come from?") | LOW | Store source URL, fetch timestamp, and generation assumptions in generated artifact metadata. |
| Context-aware question presets by doc type | Faster onboarding for API docs vs tutorials vs reference pages | MEDIUM | Infer page type and choose targeted prompts to reduce question count and improve precision. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Fully automatic install with no approval | Feels faster and "hands-off" | High risk of polluting or breaking local skill registry; violates user-control expectations | Mandatory approval gate plus preview/diff before write |
| Multi-page/domain crawling in v1 | Users want "complete skill from whole docs site" | Large ambiguity, token cost, and extraction variance reduce reliability for first release | Keep v1 single-page deterministic; add curated multi-page ingestion later |
| "One-shot no-questions" generation mode as default | Looks simple in demos | Produces brittle skills that miss boundaries, conflicts, and install constraints | Keep interactive adaptive Q&A as default; optional fast mode behind explicit flag |
| Auto-merge on detected overlap | Reduces prompts in happy path | Silent semantic drift and accidental behavior changes to trusted skills | Require merge plan review and explicit confirmation of merge strategy |
| Aggressive remote telemetry by default | Useful for analytics and tuning | Misaligned with local-dev trust model and single-user first scope | Local-only logs by default; explicit opt-in diagnostics export |

## Feature Dependencies

```text
[URL ingestion + extraction]
    └──requires──> [Content normalization]
                        └──requires──> [Schema/quality validator]
                                             └──drives──> [Adaptive Q&A loop]
                                                              └──enables──> [Final generation]

[Existing-skill index + overlap detection]
    └──requires──> [Skill parser + semantic matcher]
                        └──enables──> [Merge/update planner]
                                             └──gates──> [Install approval]

[Preview/diff]
    └──enhances──> [Install approval]

[Session persistence]
    └──enhances──> [Adaptive Q&A loop]

[Multi-page crawling]
    └──conflicts (v1 reliability/scope)──> [Deterministic single-URL ingestion]
```

### Dependency Notes

- **Adaptive Q&A requires schema + quality validation:** you need explicit failed checks to generate targeted follow-up questions.
- **Merge/update handling requires overlap detection first:** no safe merge path exists without both name-level and semantic conflict signals.
- **Approval gate depends on preview/diff quality:** users can only provide meaningful consent when they can inspect exact changes.
- **Session resume depends on persisted intermediate model state:** to avoid re-asking already answered questions.
- **Single-page v1 scope conflicts with crawling:** crawl breadth changes extraction determinism and dramatically increases conflict surface.

## MVP Definition

### Launch With (v1)

Minimum viable product - what's needed to validate the concept.

- [ ] Single URL ingest + robust content extraction - foundation for all downstream generation quality.
- [ ] Adaptive interactive Q&A tied to strict validator - ensures required skill sections are complete and usable.
- [ ] Overlap detection + explicit merge/update choice - avoids duplicate/conflicting installed skills.
- [ ] Approval-gated install to `$CODEX_HOME/skills` with preview - preserves user control and trust.
- [ ] Fail-closed behavior (no install on validation/merge uncertainty) - prioritizes reliability over throughput.

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] Skill atomicity scorer and split recommendations - add when users create too-broad generated skills.
- [ ] Install simulation + rollback helper - add when users request safer upgrade workflows.
- [ ] Doc-type-aware question presets - add when question fatigue appears in usage feedback.
- [ ] Better semantic overlap engine (embeddings/classifier hybrid) - add when false positives/negatives become operational pain.

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Controlled multi-page ingestion - defer until deterministic chunking and confidence scoring are proven.
- [ ] Team/shared governance features - defer due to single-user-first constraint.
- [ ] Optional IDE companion UX - defer until core CLI flow quality is stable.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Single URL ingest + extraction | HIGH | MEDIUM | P1 |
| Adaptive Q&A + strict validation gate | HIGH | HIGH | P1 |
| Overlap detection + merge/update choice | HIGH | HIGH | P1 |
| Explicit approval + preview install | HIGH | LOW | P1 |
| Session resume/refinement | MEDIUM | MEDIUM | P2 |
| Atomicity scoring | MEDIUM | MEDIUM | P2 |
| Install simulation + rollback | MEDIUM | MEDIUM | P2 |
| Multi-page ingestion | MEDIUM | HIGH | P3 |
| Team governance/workflows | LOW (for v1 persona) | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Codex CLI | Claude Code | Aider | Skill Weaver Approach |
|---------|-----------|-------------|-------|-----------------------|
| Permission/approval gating | Explicit approval + sandbox modes | Fine-grained permission modes and rules | User confirms edits in workflow | Mandatory approval before install to skill registry |
| Session continuation | Resume/fork support | Continue/resume support | Conversational iterative loop | Persist Q&A state + resumable refinement sessions |
| Structured output guardrails | JSON/schema-oriented automation flags | JSON schema in print mode | Diff/test/lint guardrails | Strict skill schema + quality gate before write |
| Extensibility pattern | MCP/tool integrations | Skills + hooks + subagents | Commands + scripts + model flexibility | URL-to-skill pipeline with conflict-aware install semantics |
| URL content handling | Web search/fetch integrations | WebFetch + hooks ecosystem | URL/web page support in usage modes | Deterministic single-page extraction + adaptive clarification |

## Sources

- [OpenAI Codex CLI Reference](https://developers.openai.com/codex/cli/reference/) (HIGH)
- [OpenAI Codex CLI Overview](https://help.openai.com/en/articles/11096431) (HIGH)
- [Claude Code Permissions](https://code.claude.com/docs/en/permissions) (HIGH)
- [Claude Code CLI Reference](https://code.claude.com/docs/en/cli-reference) (HIGH)
- [Claude Code Skills/Slash Commands](https://code.claude.com/docs/en/slash-commands) (HIGH)
- [Claude Code Hooks](https://code.claude.com/docs/en/hooks) (HIGH)
- [Aider Usage](https://aider.chat/docs/usage.html) (HIGH)
- [Aider Linting and Testing](https://aider.chat/docs/usage/lint-test.html) (HIGH)
- [Firecrawl Scrape Feature Docs](https://docs.firecrawl.dev/features/scrape) (MEDIUM-HIGH)
- [Jina Reader Docs](https://r.jina.ai/docs) (MEDIUM)

---
*Feature research for: CLI skill generator (Skill Weaver)*
*Researched: 2026-03-10*
