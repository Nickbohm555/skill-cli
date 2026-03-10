0b. Study @IMPLEMENTATION_PLAN.md.
0c. For reference, the application source code is in `cmd/` and supporting packages in `internal/` (unless specified otherwise).

997. Iteration scope: complete the item specified at the top of the plan.

1. Take the item in @IMPLEMENTATION_PLAN.md where it says: "Current section to work on:"

2. Before making changes, search the codebase and search reference files in your section to see how it is working. Be thorough, take your time in understanding. If parts can be re-used, do it.

3. **Run scope enforcement** (must follow):
   - One run does exactly one **Execution** OR one **Verification** session, never both.
   - Follow the section’s instructions verbatim; do not advance extra tasks.

4. After implementing functionality, run the required verification steps from the plan section. Use Go-native commands (check @AGENTS.md) as appropriate. Capture any relevant output.

5. When the task is completed: copy that item from @IMPLEMENTATION_PLAN.md and append it to @completed.md. Include some notes about what you did / what was blockig you if anything. Move the specified item to work on +1 for the next turn.

6. After completion or blocked, write `.loop-commit-msg` and end this run. For `.loop-commit-msg`, add a short summary of what was built and tested.

7. **Roadmap/Summary updates guardrail**:
   - Update `*-SUMMARY.md` only after a **plan document** within a phase is fully completed.
   - Update `ROADMAP.md` only after an **entire phase** is completed.
   - Before either update, check `@completed.md` to confirm the relevant task execution + verification sections were completed. If the required sections are not recorded in `@completed.md`, stop and ask for clarification.

NOTE: Keep @AGENTS.md operational only (how to build/test/run). Keep remaining work in @IMPLEMENTATION_PLAN.md; record completed items in @completed.md.
NOTE: Prefer complete functionality over placeholders/stubs unless explicitly needed.
