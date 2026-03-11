package overlap

import (
	"errors"
	"io"
	"testing"
	"time"
)

func TestDecisionFlowAcceptsExplicitChoice(t *testing.T) {
	t.Parallel()

	var captured DecisionPrompt
	now := time.Date(2026, time.March, 11, 18, 0, 0, 0, time.UTC)
	flow := NewDecisionFlow(DecisionPrompterFunc(func(prompt DecisionPrompt) (ResolutionMode, error) {
		captured = prompt
		return ResolutionMerge, nil
	}))
	flow.Now = func() time.Time { return now }

	report := OverlapReport{
		Candidate:       SkillProfile{ID: "candidate.docs"},
		OverallSeverity: SeverityHigh,
		Findings: []OverlapFinding{
			{
				RuleID:          "OVLP.NAME.EXACT",
				ExistingSkillID: "installed.docs",
				Severity:        SeverityHigh,
				Explanation:     "Candidate name exactly matches an installed skill name.",
				ExplanationMeta: ExplanationMetadata{
					Summary: "Exact collision on skill name.",
					RuleIDs: []string{"OVLP.NAME.EXACT"},
				},
			},
		},
	}

	updated, message := flow.Decide(report)

	if captured.TargetSkillID != "installed.docs" {
		t.Fatalf("prompt target = %q, want %q", captured.TargetSkillID, "installed.docs")
	}
	if len(captured.Options) != 3 {
		t.Fatalf("prompt options len = %d, want 3", len(captured.Options))
	}
	gotModes := []ResolutionMode{captured.Options[0].Mode, captured.Options[1].Mode, captured.Options[2].Mode}
	wantModes := []ResolutionMode{ResolutionUpdate, ResolutionMerge, ResolutionAbort}
	for i := range wantModes {
		if gotModes[i] != wantModes[i] {
			t.Fatalf("option mode[%d] = %q, want %q", i, gotModes[i], wantModes[i])
		}
	}
	if updated.Decision == nil {
		t.Fatal("decision = nil, want explicit resolution decision")
	}
	if updated.Decision.Mode != ResolutionMerge {
		t.Fatalf("decision mode = %q, want %q", updated.Decision.Mode, ResolutionMerge)
	}
	if updated.Decision.Blocking {
		t.Fatal("decision blocking = true, want false")
	}
	if updated.Decision.TargetSkillID != "installed.docs" {
		t.Fatalf("decision target = %q, want %q", updated.Decision.TargetSkillID, "installed.docs")
	}
	if updated.Decision.SelectedAt == nil || !updated.Decision.SelectedAt.Equal(now) {
		t.Fatalf("decision selected_at = %#v, want %v", updated.Decision.SelectedAt, now)
	}
	if message != `Selected conflict resolution: merge with existing skill "installed.docs".` {
		t.Fatalf("message = %q", message)
	}
}

func TestDecisionFlowShortCircuitsWhenNoOverlapExists(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 11, 18, 5, 0, 0, time.UTC)
	flow := NewDecisionFlow(nil)
	flow.Now = func() time.Time { return now }

	updated, message := flow.Decide(OverlapReport{
		Candidate:       SkillProfile{ID: "candidate.docs"},
		OverallSeverity: SeverityNone,
	})

	if updated.Decision == nil {
		t.Fatal("decision = nil, want new-install shortcut decision")
	}
	if updated.Decision.Mode != ResolutionNewInstall {
		t.Fatalf("decision mode = %q, want %q", updated.Decision.Mode, ResolutionNewInstall)
	}
	if updated.Decision.Blocking {
		t.Fatal("decision blocking = true, want false")
	}
	if updated.Decision.SelectedAt == nil || !updated.Decision.SelectedAt.Equal(now) {
		t.Fatalf("decision selected_at = %#v, want %v", updated.Decision.SelectedAt, now)
	}
	if message != "No conflicts found. Proceeding with the new-install path." {
		t.Fatalf("message = %q", message)
	}
}

func TestDecisionFlowFallsBackToBlockingAbortOnInterruptedPrompt(t *testing.T) {
	t.Parallel()

	flow := NewDecisionFlow(DecisionPrompterFunc(func(prompt DecisionPrompt) (ResolutionMode, error) {
		return "", io.EOF
	}))

	updated, message := flow.Decide(OverlapReport{
		Candidate:       SkillProfile{ID: "candidate.docs"},
		OverallSeverity: SeverityMedium,
		Findings: []OverlapFinding{
			{
				RuleID:          "OVLP.STRUCTURAL.OVERLAP",
				ExistingSkillID: "installed.docs",
				Severity:        SeverityMedium,
				Explanation:     "Candidate moderately overlaps an installed skill and requires explicit resolution.",
			},
		},
	})

	if updated.Decision == nil {
		t.Fatal("decision = nil, want blocking abort")
	}
	if updated.Decision.Mode != ResolutionAbort {
		t.Fatalf("decision mode = %q, want %q", updated.Decision.Mode, ResolutionAbort)
	}
	if !updated.Decision.Blocking {
		t.Fatal("decision blocking = false, want true")
	}
	if updated.Decision.SelectedAt != nil {
		t.Fatalf("decision selected_at = %#v, want nil", updated.Decision.SelectedAt)
	}
	if updated.Decision.TargetSkillID != "installed.docs" {
		t.Fatalf("decision target = %q, want %q", updated.Decision.TargetSkillID, "installed.docs")
	}
	if message != "Conflict resolution was interrupted. Defaulting to abort and keeping install handoff blocked." {
		t.Fatalf("message = %q", message)
	}
}

func TestDecisionFlowFallsBackToBlockingAbortOnInvalidSelection(t *testing.T) {
	t.Parallel()

	flow := NewDecisionFlow(DecisionPrompterFunc(func(prompt DecisionPrompt) (ResolutionMode, error) {
		return ResolutionNewInstall, nil
	}))

	updated, _ := flow.Decide(OverlapReport{
		Candidate:       SkillProfile{ID: "candidate.docs"},
		OverallSeverity: SeverityHigh,
		Findings: []OverlapFinding{
			{
				RuleID:          "OVLP.NAME.EXACT",
				ExistingSkillID: "installed.docs",
				Severity:        SeverityHigh,
				Explanation:     "Candidate name exactly matches an installed skill name.",
			},
		},
	})

	if updated.Decision == nil {
		t.Fatal("decision = nil, want blocking abort")
	}
	if updated.Decision.Mode != ResolutionAbort || !updated.Decision.Blocking {
		t.Fatalf("decision = %#v, want blocking abort", *updated.Decision)
	}
}

func TestInterruptedDecisionExplanationIncludesPromptError(t *testing.T) {
	t.Parallel()

	msg := interruptedDecisionExplanation(errors.New("ctrl+c"))
	if msg == "" {
		t.Fatal("interruptedDecisionExplanation() = empty, want explanation")
	}
}
