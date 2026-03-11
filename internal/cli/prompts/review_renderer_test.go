package prompts

import (
	"strings"
	"testing"

	"github.com/Nickbohm555/skill-cli/internal/refinement"
)

func TestBuildReviewModelGroupsSectionsAndReadiness(t *testing.T) {
	t.Parallel()

	report := reviewReportFromState(t, reviewState(t, map[refinement.FieldID]string{
		refinement.FieldOutOfScope: "",
	}))

	model := BuildReviewModel(report)
	if model.CommitReady {
		t.Fatal("BuildReviewModel() CommitReady = true, want false")
	}
	if len(model.Sections) != 4 {
		t.Fatalf("BuildReviewModel() sections = %d, want 4", len(model.Sections))
	}
	if model.Sections[0].Title != "Purpose" {
		t.Fatalf("first section title = %q, want %q", model.Sections[0].Title, "Purpose")
	}
	if model.Sections[3].Title != "Boundaries" {
		t.Fatalf("last section title = %q, want %q", model.Sections[3].Title, "Boundaries")
	}

	outOfScope := findReviewField(t, model, refinement.FieldOutOfScope)
	if outOfScope.StatusLabel != "missing" {
		t.Fatalf("FieldOutOfScope status = %q, want %q", outOfScope.StatusLabel, "missing")
	}
	if outOfScope.Answer != "(no answer yet)" {
		t.Fatalf("FieldOutOfScope answer = %q, want placeholder", outOfScope.Answer)
	}
	if len(outOfScope.Hints) != 1 || outOfScope.Hints[0] != "Add an answer before commit." {
		t.Fatalf("FieldOutOfScope hints = %v, want required-missing hint", outOfScope.Hints)
	}
}

func TestRenderReviewIncludesGroupedSectionsStatusesAndRevisionHints(t *testing.T) {
	t.Parallel()

	state := reviewState(t, nil)
	if _, err := state.ReviseAnswer(
		refinement.FieldPurposeSummary,
		"Refocus the skill on database docs only, excluding generic examples.",
		refinement.DefaultFieldGraph(),
	); err != nil {
		t.Fatalf("ReviseAnswer() error = %v", err)
	}

	rendered := RenderReview(reviewReportFromState(t, state))

	wantSubstrings := []string{
		"Commit readiness: blocked",
		"Purpose [needs attention]",
		"Constraints [ready]",
		"Examples [needs attention]",
		"Boundaries [needs attention]",
		"- Purpose Summary [needs attention]",
		"Hint: Re-opened because a related answer changed.",
	}

	for _, want := range wantSubstrings {
		if !strings.Contains(rendered, want) {
			t.Fatalf("RenderReview() missing %q\nrendered:\n%s", want, rendered)
		}
	}
}

func TestRenderReviewShowsReadySummaryWhenCommitGatePasses(t *testing.T) {
	t.Parallel()

	rendered := RenderReview(reviewReportFromState(t, reviewState(t, nil)))
	if !strings.Contains(rendered, "Commit readiness: ready") {
		t.Fatalf("RenderReview() missing ready banner\nrendered:\n%s", rendered)
	}
	if !strings.Contains(rendered, "All required fields are ready. You can commit this refinement session.") {
		t.Fatalf("RenderReview() missing ready summary\nrendered:\n%s", rendered)
	}
}

func reviewReportFromState(t *testing.T, state *refinement.SessionState) refinement.ValidationReport {
	t.Helper()

	report, err := refinement.DefaultValidator().Evaluate(state)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	return report
}

func reviewState(t *testing.T, overrides map[refinement.FieldID]string) *refinement.SessionState {
	t.Helper()

	state, err := refinement.NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	answers := map[refinement.FieldID]string{
		refinement.FieldPurposeSummary:  "Generate a Codex skill from one docs URL, including install steps, scope boundaries, and review-ready examples.",
		refinement.FieldPrimaryTasks:    "Capture the docs source, extract implementation guidance, and turn it into a focused skill with explicit installation steps.",
		refinement.FieldSuccessCriteria: "The generated skill is installable, scoped to one domain, and includes concrete usage examples plus constraints.",
		refinement.FieldConstraints:     "Use one docs URL only, keep the skill deterministic, and exclude unsupported setup steps or speculative workflows.",
		refinement.FieldDependencies:    "Requires network access for docs fetches, a reachable documentation site, and OpenAI credentials only when structured summarization is enabled.",
		refinement.FieldExampleRequests: "Examples should include generating a skill from Go docs and refining boundaries when the docs mix tutorials with API references.",
		refinement.FieldExampleOutputs:  "Output examples must show install commands, supported inputs, and one explicit out-of-scope case for the final skill.",
		refinement.FieldInScope:         "In scope: extracting skill instructions, installation notes, supported commands, and concrete examples from the chosen docs set.",
		refinement.FieldOutOfScope:      "Out of scope: building unrelated tooling, inventing missing APIs, or merging content from multiple unrelated documentation sites.",
	}

	for fieldID, override := range overrides {
		answers[fieldID] = override
	}

	for _, fieldID := range state.RequiredFields() {
		if err := state.SetAnswer(fieldID, answers[fieldID]); err != nil {
			t.Fatalf("SetAnswer(%q) error = %v", fieldID, err)
		}
		if strings.TrimSpace(answers[fieldID]) == "" {
			continue
		}
		if err := state.MarkReady(fieldID); err != nil {
			t.Fatalf("MarkReady(%q) error = %v", fieldID, err)
		}
	}

	return state
}

func findReviewField(t *testing.T, model ReviewModel, fieldID refinement.FieldID) ReviewField {
	t.Helper()

	for _, section := range model.Sections {
		for _, field := range section.Fields {
			if field.ID == fieldID {
				return field
			}
		}
	}

	t.Fatalf("field %q not found in review model", fieldID)
	return ReviewField{}
}
