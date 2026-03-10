# Roadmap: Skill Weaver

## Overview

This roadmap delivers Skill Weaver as a quality-gated, local-first CLI that turns documentation into installable Codex skills with explicit user control. Phases are derived directly from v1 requirement groups and ordered by dependency so each phase unlocks the next capability. Every v1 requirement is mapped to exactly one phase with observable outcomes.

## Phases

### Phase 1 - Crawl & Ingestion Foundation

**Goal:** User can start from a documentation entry URL and get a bounded, transparent crawl result suitable for downstream processing.

**Dependencies:** None

**Requirements:**
- CRAWL-01
- CRAWL-02
- CRAWL-03
- CRAWL-04

**Success Criteria:**
1. User can input a docs URL and the system crawls only same-domain pages.
2. Crawl stops safely at the default 50-page cap without hanging or overrunning.
3. User can see which URLs were skipped and why.
4. User receives a final crawl summary with discovered, processed, and skipped counts.

### Phase 2 - Content Processing & Attribution

**Goal:** User gets clean, attributable source material that is efficient to use in generation.

**Dependencies:** Phase 1

**Requirements:**
- CONT-01
- CONT-02
- CONT-03

**Success Criteria:**
1. Crawled pages are converted into normalized text that is usable by later pipeline steps.
2. User can inspect structured summarized chunks rather than raw unbounded page dumps.
3. Generated context retains source URL attribution that the user can review.

### Phase 3 - Interactive Refinement Loop

**Goal:** User can complete and refine required skill inputs through an adaptive conversation before generation finalization.

**Dependencies:** Phase 2

**Requirements:**
- INT-01
- INT-02
- INT-03

**Success Criteria:**
1. User is guided through adaptive questions until required skill fields are filled.
2. Flow behavior reflects a crawl/summarize-first approach followed by targeted deepening questions.
3. User can revise prior answers before committing to final generation output.

### Phase 4 - Validation & Quality Gates

**Goal:** User can only proceed when generated skill output is structurally and semantically acceptable for Codex usage.

**Dependencies:** Phase 3

**Requirements:**
- VAL-01
- VAL-02
- VAL-03

**Success Criteria:**
1. Candidate skill fails closed when required sections or schema constraints are not met.
2. On validation failure, user receives targeted follow-up prompts instead of a silent error.
3. Final candidate explicitly shows in-scope and out-of-scope boundaries.

### Phase 5 - Overlap & Conflict Resolution

**Goal:** User can resolve conflicts with existing installed skills explicitly before any install decision.

**Dependencies:** Phase 4

**Requirements:**
- OVLP-01
- OVLP-02
- OVLP-03

**Success Criteria:**
1. System checks generated skill against installed skills and surfaces overlap/conflict findings.
2. When overlap exists, user is offered explicit merge or update choices.
3. User sees the selected conflict-resolution outcome before any write/install action.

### Phase 6 - Approval-Gated Install & Activation

**Goal:** User can safely approve and install a validated, conflict-resolved skill that Codex can use immediately.

**Dependencies:** Phase 5

**Requirements:**
- INST-01
- INST-02
- INST-03
- INST-04

**Success Criteria:**
1. User can review a preview or diff of generated skill content before approval.
2. Skill is written to `$CODEX_HOME/skills/<skill-name>` only after explicit user approval.
3. Install is blocked when validation or conflict state remains unresolved.
4. After successful install, Codex can use the skill without additional manual setup.

## Progress

| Phase | Name | Status | Requirements |
|------|------|--------|--------------|
| 1 | Crawl & Ingestion Foundation | Pending | 4 |
| 2 | Content Processing & Attribution | Pending | 3 |
| 3 | Interactive Refinement Loop | Pending | 3 |
| 4 | Validation & Quality Gates | Pending | 3 |
| 5 | Overlap & Conflict Resolution | Pending | 3 |
| 6 | Approval-Gated Install & Activation | Pending | 4 |

## Coverage Check

- Total v1 requirements: 20
- Requirements mapped: 20
- Unmapped requirements: 0
- Duplicate mappings: 0

Coverage is complete.
