package refinement

import "testing"

func TestClarityAssessmentHighSpecificityPasses(t *testing.T) {
	t.Parallel()

	policy := DefaultClarityPolicy()
	assessment, err := policy.Assess(FieldPurposeSummary, "Generate a Codex skill from one docs URL, including install steps, scope boundaries, and example requests for later review.")
	if err != nil {
		t.Fatalf("Assess() error = %v", err)
	}

	if !assessment.Pass {
		t.Fatalf("Assess().Pass = false, want true: %+v", assessment)
	}
	if assessment.Score < assessment.Threshold {
		t.Fatalf("Assess().Score = %d, threshold = %d", assessment.Score, assessment.Threshold)
	}
	if assessment.WordCount < 10 {
		t.Fatalf("Assess().WordCount = %d, want >= 10", assessment.WordCount)
	}
}

func TestClarityAssessmentShortAmbiguousFails(t *testing.T) {
	t.Parallel()

	policy := DefaultClarityPolicy()
	assessment, err := policy.Assess(FieldConstraints, "Maybe keep it flexible and stuff.")
	if err != nil {
		t.Fatalf("Assess() error = %v", err)
	}

	if assessment.Pass {
		t.Fatalf("Assess().Pass = true, want false: %+v", assessment)
	}
	if assessment.Score >= assessment.Threshold {
		t.Fatalf("Assess().Score = %d, threshold = %d", assessment.Score, assessment.Threshold)
	}
	if len(assessment.Penalties) == 0 {
		t.Fatalf("Assess().Penalties = %v, want ambiguity penalty", assessment.Penalties)
	}
}

func TestClarityAssessmentStructuredExamplePasses(t *testing.T) {
	t.Parallel()

	policy := DefaultClarityPolicy()
	assessment, err := policy.Assess(FieldExampleRequests, "Examples: `cli-skill process --url https://go.dev/doc/`; ask for install commands; include one failure case.")
	if err != nil {
		t.Fatalf("Assess() error = %v", err)
	}

	if !assessment.Pass {
		t.Fatalf("Assess().Pass = false, want true: %+v", assessment)
	}
	if assessment.Score < assessment.Threshold {
		t.Fatalf("Assess().Score = %d, threshold = %d", assessment.Score, assessment.Threshold)
	}
}

func TestClarityDeepeningDecisionEscalatesAndCaps(t *testing.T) {
	t.Parallel()

	policy := DefaultClarityPolicy()
	answer := "Things maybe."

	first, err := policy.DeepeningDecision(FieldOutOfScope, answer, 0)
	if err != nil {
		t.Fatalf("DeepeningDecision(first) error = %v", err)
	}
	if first.Mode != DeepeningModeFreeText {
		t.Fatalf("DeepeningDecision(first).Mode = %q, want %q", first.Mode, DeepeningModeFreeText)
	}
	if first.RequireExplicitOther {
		t.Fatalf("DeepeningDecision(first).RequireExplicitOther = true, want false")
	}

	second, err := policy.DeepeningDecision(FieldOutOfScope, answer, 1)
	if err != nil {
		t.Fatalf("DeepeningDecision(second) error = %v", err)
	}
	if second.Mode != DeepeningModeStructuredChoice {
		t.Fatalf("DeepeningDecision(second).Mode = %q, want %q", second.Mode, DeepeningModeStructuredChoice)
	}
	if !second.RequireExplicitOther {
		t.Fatalf("DeepeningDecision(second).RequireExplicitOther = false, want true")
	}

	capped, err := policy.DeepeningDecision(FieldOutOfScope, answer, 2)
	if err != nil {
		t.Fatalf("DeepeningDecision(capped) error = %v", err)
	}
	if capped.Mode != DeepeningModeCapped {
		t.Fatalf("DeepeningDecision(capped).Mode = %q, want %q", capped.Mode, DeepeningModeCapped)
	}
}

func TestClarityDeepeningDecisionStopsForClearAnswer(t *testing.T) {
	t.Parallel()

	policy := DefaultClarityPolicy()
	decision, err := policy.DeepeningDecision(FieldPrimaryTasks, "Capture the docs URL, extract usage guidance, and generate a focused Codex skill with explicit installation notes.", 0)
	if err != nil {
		t.Fatalf("DeepeningDecision() error = %v", err)
	}

	if decision.Mode != DeepeningModeNone {
		t.Fatalf("DeepeningDecision().Mode = %q, want %q", decision.Mode, DeepeningModeNone)
	}
}
