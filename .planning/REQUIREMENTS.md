# Requirements: Skill Weaver

**Defined:** 2026-03-10
**Core Value:** Generate a skill that is actually usable in Codex, with clear scope and correct installation, in one guided flow.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Crawl & Ingestion

- [ ] **CRAWL-01**: User can provide a documentation entry URL and the system crawls same-domain pages only.
- [ ] **CRAWL-02**: Crawl runs with a default cap of 50 pages and stops safely when the cap is reached.
- [ ] **CRAWL-03**: System skips unsupported or low-signal pages and reports skipped URLs with reasons.
- [ ] **CRAWL-04**: System shows a crawl summary with discovered, processed, and skipped page counts.

### Content Processing

- [ ] **CONT-01**: System extracts normalized text from crawled pages for downstream generation.
- [ ] **CONT-02**: System summarizes extracted content into structured chunks for prompt efficiency.
- [ ] **CONT-03**: System preserves source URL attribution for generated skill context.

### Interactive Refinement

- [ ] **INT-01**: User is guided through an adaptive interactive question flow to complete required skill fields.
- [ ] **INT-02**: The flow uses a hybrid strategy: crawl and summarize first, then ask targeted deepening questions.
- [ ] **INT-03**: User can revise responses before final skill generation.

### Validation & Quality Gates

- [ ] **VAL-01**: Generated skill must pass strict required-section and schema validation before install is allowed.
- [ ] **VAL-02**: If validation fails, install is blocked and targeted follow-up questions are asked.
- [ ] **VAL-03**: Generated skill must define explicit in-scope and out-of-scope boundaries.

### Overlap & Conflict Handling

- [ ] **OVLP-01**: System checks generated skill against installed skills for overlap/conflict.
- [ ] **OVLP-02**: If overlap is detected, user is offered explicit merge or update paths.
- [ ] **OVLP-03**: Conflict resolution outcome is shown before any write operation.

### Install & Safety

- [ ] **INST-01**: User sees a preview or diff of generated skill content before approval.
- [ ] **INST-02**: Skill is written to `$CODEX_HOME/skills/<skill-name>` only after explicit user approval.
- [ ] **INST-03**: Install is fail-closed when validation or conflict state is unresolved.
- [ ] **INST-04**: Installed skill is immediately usable by Codex.

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Cost & Retrieval

- **COST-01**: System adds relevance-based page selection/indexing in addition to summarization.

### Workflow UX

- **UX-01**: User can resume an interrupted refinement session without restarting.

### Install Safety

- **SAFE-01**: System offers install simulation and rollback helper commands.

### Skill Quality

- **QUAL-01**: System scores skill atomicity and suggests splitting overly broad outputs.

### Crawl Expansion

- **CRAWL-05**: User can configure higher crawl limits and advanced crawl policies.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Multi-domain crawling | Expands complexity and ambiguity beyond v1 reliability goals |
| Unlimited crawl depth/page count | Creates unpredictable token/time cost and weakens determinism |
| Team/shared governance workflows | Initial target is single-user workflow |
| Auto-install without approval | Violates explicit user-control requirement |
| One-shot default mode without interactive refinement | Conflicts with quality-first adaptive flow |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CRAWL-01 | TBD | Pending |
| CRAWL-02 | TBD | Pending |
| CRAWL-03 | TBD | Pending |
| CRAWL-04 | TBD | Pending |
| CONT-01 | TBD | Pending |
| CONT-02 | TBD | Pending |
| CONT-03 | TBD | Pending |
| INT-01 | TBD | Pending |
| INT-02 | TBD | Pending |
| INT-03 | TBD | Pending |
| VAL-01 | TBD | Pending |
| VAL-02 | TBD | Pending |
| VAL-03 | TBD | Pending |
| OVLP-01 | TBD | Pending |
| OVLP-02 | TBD | Pending |
| OVLP-03 | TBD | Pending |
| INST-01 | TBD | Pending |
| INST-02 | TBD | Pending |
| INST-03 | TBD | Pending |
| INST-04 | TBD | Pending |

**Coverage:**
- v1 requirements: 20 total
- Mapped to phases: 0
- Unmapped: 20 ⚠️

---
*Requirements defined: 2026-03-10*
*Last updated: 2026-03-10 after initial definition*
