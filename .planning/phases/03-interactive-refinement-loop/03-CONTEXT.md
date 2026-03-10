# Phase 3: Interactive Refinement Loop - Context

**Gathered:** 2026-03-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Enable an adaptive conversation that collects and refines all required skill inputs before generation finalization, including the ability to revise prior answers before commit. This phase does not add new capabilities beyond the refinement loop itself.

</domain>

<decisions>
## Implementation Decisions

### Adaptive questioning flow
- Start in domain chunks (for example: purpose, inputs, constraints, examples), with depth within each chunk.
- Use confidence-driven branching: ask deeper follow-ups when answers are vague or ambiguous.
- For unclear answers, use concrete multiple-choice clarifications with an "other" path.
- Move to review only when required fields reach a higher quality/clarity bar, not just minimum completion.

### Answer revision experience
- Revision entry uses command-style interaction (`revise <field>`).
- Editing a field auto re-asks directly affected follow-up questions.
- Show only the latest answer during revision (no visible history trail).
- After saving, show a brief summary of what changed and what may be impacted.

### Final review and commit
- Final review is grouped by sections (purpose, constraints, examples, boundaries), not a flat list.
- Show clear readiness indicators per field (ready vs needs attention).
- Block final commit when required fields are missing or low-clarity.
- After commit confirmation, lock answers and proceed directly to generation.

### Claude's Discretion
- Exact wording and visual style of prompts and readiness indicators.
- How confidence/clarity thresholds are operationalized internally.
- Exact formatting of sectioned review output.

</decisions>

<specifics>
## Specific Ideas

No external product references were specified. Preferences emphasize adaptive depth, command-driven revision, and strict commit gating on required-field clarity.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 03-interactive-refinement-loop*
*Context gathered: 2026-03-10*
