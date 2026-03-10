# Pitfalls Research

**Domain:** CLI that generates Codex `SKILL.md` artifacts from single-page docs URLs
**Researched:** 2026-03-10
**Confidence:** MEDIUM-HIGH

## Critical Pitfalls

### Pitfall 1: Treating fetched docs content as trusted instructions

**What goes wrong:**
Malicious or accidental instructions inside fetched page content influence generation output (scope drift, unsafe commands, hidden directives), producing unsafe or misleading skills.

**Why it happens:**
The pipeline conflates "source data" and "instructions to follow" during extraction/generation, and does not model remote content as untrusted input.

**How to avoid:**
- Treat fetched page text as untrusted data by default.
- Sanitize/normalize HTML and strip hidden or suspicious payload patterns before generation.
- Use strict structured generation contracts and deterministic validators.
- Require explicit user approval before any write/install step (already aligned with project constraints).
- Add adversarial test fixtures (prompt injection strings, encoded variants, hidden text).

**Warning signs:**
- Generated skill includes imperative instructions not present in user intent.
- Scope suddenly expands ("do everything" style behavior).
- Skill output references secrets/system prompt style text.
- Minor wording changes in source page cause major behavioral changes.

**Phase to address:**
Phase 1 (Safe ingestion boundaries) and Phase 6 (Adversarial hardening tests).

---

### Pitfall 2: Assuming single-page HTML contains the real documentation

**What goes wrong:**
The generator builds skills from incomplete content because the target docs page is JS-rendered, gated, or missing critical sections in the raw HTML response.

**Why it happens:**
v1 ingestion is intentionally one URL; teams overestimate what plain HTTP+HTML parsing can capture on modern docs stacks.

**How to avoid:**
- Add extraction completeness checks (word count, heading density, required-section heuristics).
- Detect low-signal pages and trigger user prompts ("content appears incomplete; proceed or provide another URL").
- Persist extraction diagnostics in preview output.
- Keep parser behavior explicit: static HTML mode only in v1, with clear UX messaging.

**Warning signs:**
- Extracted text is mostly nav/footer boilerplate.
- Missing examples/code blocks despite known docs richness.
- Frequent user corrections like "that section was not on the page."
- Large variance in generated quality across similar URLs.

**Phase to address:**
Phase 2 (Extraction quality gates and diagnostics).

---

### Pitfall 3: Weak skill-shape validation ("looks valid" but not usable)

**What goes wrong:**
Generated `SKILL.md` passes superficial checks but fails practical usage: unclear trigger boundaries, broad scope, missing constraints, or ambiguous instructions.

**Why it happens:**
Validation focuses on syntax/frontmatter presence rather than semantic quality and single-capability boundaries.

**How to avoid:**
- Validate both structure and semantics (required sections, explicit trigger boundaries, non-goals).
- Add a rubric-based quality gate before install.
- Re-ask targeted questions on specific failed dimensions (scope, activation criteria, exclusions).
- Include overlap/ambiguity checks against existing local skills.

**Warning signs:**
- Many generated skills start with generic descriptions ("helps with many tasks").
- High overlap scores between new and existing skills.
- User repeatedly edits installed skills immediately after generation.
- Trigger behavior is inconsistent across similar prompts.

**Phase to address:**
Phase 3 (Semantic validation and adaptive re-ask loop).

---

### Pitfall 4: Name collisions and overlap are handled too late

**What goes wrong:**
New skills collide with existing names/capabilities and create confusing invocation behavior, duplicate options, or hidden conflicts after installation.

**Why it happens:**
Conflict detection is deferred to install time or reduced to filename checks instead of capability-level overlap checks.

**How to avoid:**
- Run overlap and naming checks before final approval step.
- Present merge/update/rename choices with a preview diff.
- Block install when overlap exceeds threshold unless user explicitly approves override path.
- Keep deterministic conflict resolution rules.

**Warning signs:**
- Two skills with nearly identical descriptions appear in selector lists.
- Users report "wrong skill gets picked" during implicit invocation.
- Frequent post-install manual renames and deletions.

**Phase to address:**
Phase 4 (Conflict detection + explicit resolution UX).

---

### Pitfall 5: Non-atomic install writes corrupt local skill state

**What goes wrong:**
Interrupted or partial writes leave broken skill directories; Codex sees invalid or mixed artifact state.

**Why it happens:**
Direct writes into final install path without staging, fsync, or rollback strategy.

**How to avoid:**
- Write to temp staging directory, validate, then perform atomic move into target.
- Keep backup/rollback on overwrite flows.
- Fail closed: no install on any validation or file operation error.
- Emit explicit post-install verification checks.

**Warning signs:**
- Partial files or missing sections after Ctrl-C/crash.
- Intermittent "skill exists but unusable" reports.
- Manual cleanup in `$CODEX_HOME/skills` becomes common.

**Phase to address:**
Phase 5 (Transactional install and rollback).

---

### Pitfall 6: Unsafe URL ingestion (SSRF, hostile redirects, hanging requests)

**What goes wrong:**
CLI can be abused to fetch internal/private endpoints, follow dangerous redirect chains, or hang indefinitely on network operations.

**Why it happens:**
Insufficient URL allow/deny policy, redirect controls, timeout defaults, and protocol restrictions.

**How to avoid:**
- Enforce allowed schemes (`http`/`https`) and deny local/private address ranges by default.
- Set strict request timeout and max redirects; surface clear error messages.
- Validate final resolved host before processing body.
- Add request-size and content-type limits for ingestion.

**Warning signs:**
- CLI appears to "freeze" on URL fetch.
- Logs show unexpected internal network hosts.
- Large/binary payloads reach extraction stage.
- Redirect chains trigger inconsistent behavior.

**Phase to address:**
Phase 1 (Network safety + ingestion guardrails).

---

### Pitfall 7: Approval UX exists but is not script-safe

**What goes wrong:**
Interactive confirmation prompts break automation, or automation bypasses prompts unintentionally and writes without clear auditability.

**Why it happens:**
No explicit split between interactive and non-interactive modes (`--force`/CI policy), and no consistent approval semantics.

**How to avoid:**
- Define approval policy for interactive vs non-interactive contexts.
- Require explicit force flag in scripted contexts; otherwise block with actionable message.
- Standardize confirmation messaging and post-action summaries.
- Add tests for prompt behavior in TTY, CI, and piped execution modes.

**Warning signs:**
- CI jobs hang waiting for user input.
- Users are surprised by installs in script runs.
- Prompt behavior differs between environments.

**Phase to address:**
Phase 4 (Approval UX contract and automation policy tests).

---

## Technical Debt Patterns

Shortcuts that feel fast early but become expensive quickly.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Skip semantic validation; only check YAML/header shape | Faster MVP output | Low-quality skills, high rework, user distrust | Never |
| Install directly to final path without staging | Less file handling code | Corruption/partial state on failure | Never |
| Hardcode overlap logic to filename match only | Simple implementation | Capability conflicts remain unresolved | Only for first spike, not MVP |
| Add broad fallback prompts instead of targeted re-ask | Faster to implement | Non-convergent question loop, token waste | MVP only with explicit cap + telemetry |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Remote docs URL fetch | No timeout / unlimited redirects | Use bounded timeout, max redirects, size limits, and host/scheme validation |
| HTML parsing (`goquery`) | Assume browser-like DOM completeness | Treat as static parse; add completeness heuristics and user warnings |
| Codex skills install path | Overwrite existing skill state silently | Preview diff + explicit approve + transactional write |
| Existing skill discovery | Name-only duplicate checks | Compare name + description + capability overlap score |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Re-fetch + re-parse URL on every re-ask iteration | Slow loop, repeated network spikes | Cache normalized source per run with hash key | Noticeable after 3+ retries per run |
| Full-document re-validation for tiny edits | Latency increases with doc size | Incremental validation for changed sections | Large docs or long adaptive sessions |
| Unbounded prompt context growth across retries | Rising token usage and drift | Keep compact state object + bounded history window | By mid-session on verbose docs |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| No defense against indirect prompt injection | Unsafe generated instructions; policy bypass | Untrusted-content handling, sanitization, structured generation, adversarial tests |
| Accept arbitrary URL targets incl. private/internal networks | SSRF-like local environment probing | Default-deny private ranges and localhost; explicit override with warning |
| Weak output validation before install | Invalid or dangerous skills installed | Semantic validator + install block on any gate failure |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Vague validation errors ("invalid skill") | User cannot recover, abandons flow | Targeted error messages tied to missing/weak section |
| Hidden conflict resolution | Surprise behavior after install | Show overlap rationale and explicit choice (merge/update/rename/cancel) |
| No extraction quality preview | User approves low-quality output unknowingly | Preview extracted summary and key diagnostics before generation |

## "Looks Done But Isn't" Checklist

- [ ] **Ingestion safety:** URL policy blocks private/local targets and enforces timeout/redirect limits.
- [ ] **Extraction quality:** System detects low-signal pages and asks for confirmation/alternate URL.
- [ ] **Validation rigor:** Semantic quality rubric gates install, not just syntax checks.
- [ ] **Conflict handling:** Overlap and naming conflicts require explicit user decision.
- [ ] **Install durability:** Staged write + atomic move + rollback path verified.
- [ ] **Approval semantics:** Interactive and non-interactive behavior tested and documented.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Corrupt/partial install | MEDIUM | Detect on startup, restore from backup, rerun install from staged artifact |
| Bad overlap merge | MEDIUM | Roll back merged artifact, re-run with explicit rename/update path |
| Unsafe generated content installed | HIGH | Quarantine skill, audit source URL + generation trace, strengthen sanitizer/tests before re-enable |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Untrusted-content injection | Phase 1 + Phase 6 | Red-team fixtures fail closed; no unsafe install path |
| Incomplete extraction from dynamic docs | Phase 2 | Completeness metrics + user confirmation on low-signal pages |
| Weak semantic validation | Phase 3 | Quality rubric pass required before install |
| Late collision/overlap handling | Phase 4 | Conflicts surfaced pre-approval with deterministic resolution |
| Non-atomic install writes | Phase 5 | Crash/interruption tests preserve prior valid state |
| Script/CI approval mismatch | Phase 4 | TTY/CI test matrix verifies consistent approval semantics |

## Sources

- [OWASP LLM Prompt Injection Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/LLM_Prompt_Injection_Prevention_Cheat_Sheet.html) (HIGH)
- [OWASP LLM01:2025 Prompt Injection](https://genai.owasp.org/llmrisk/llm01-prompt-injection/) (HIGH)
- [OpenAI Codex Agent Skills documentation](https://developers.openai.com/codex/skills) (HIGH)
- [RFC 9309: Robots Exclusion Protocol](https://www.rfc-editor.org/rfc/rfc9309) (HIGH)
- [Go `net/http` Client docs](https://pkg.go.dev/net/http#Client) (HIGH)
- [Goquery package documentation](https://pkg.go.dev/github.com/PuerkitoBio/goquery) (HIGH)
- [Netlify CLI PR adding confirmation prompts for unsafe commands](https://github.com/netlify/cli/pull/6878) (MEDIUM; real-world evidence, not a normative standard)
- [Playwright page API docs](https://playwright.dev/docs/api/class-page#page-wait-for-load-state) (MEDIUM; supporting evidence for dynamic-content timing pitfalls)

---
*Pitfalls research for: Skill Weaver (CLI skill generator)*
*Researched: 2026-03-10*
