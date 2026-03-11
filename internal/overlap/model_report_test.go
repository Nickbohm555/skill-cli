package overlap

import (
	"encoding/json"
	"testing"
	"time"
)

func TestModelOverlapContractsJSON(t *testing.T) {
	selectedAt := time.Date(2026, time.March, 11, 15, 4, 5, 0, time.UTC)
	report := NewReport(SkillProfile{
		ID:          "candidate.docs-skill",
		Name:        "Docs Skill",
		Description: "Generates a focused skill from docs.",
		InScope:     []string{"Extract docs guidance", "Render SKILL.md"},
		OutOfScope:  []string{"Install files"},
		Commands:    []string{"cli-skill process --url https://example.com/docs"},
		SourcePath:  "/tmp/candidate/SKILL.md",
	})
	report.AddWarning(IndexWarning{
		Path:    "/Users/nick/.codex/skills/broken/SKILL.md",
		Message: "parse failed",
	})
	report.AddFinding(OverlapFinding{
		RuleID:          "OVLP.NAME.EXACT",
		ExistingSkillID: "installed.docs-skill",
		Severity:        SeverityHigh,
		Score:           1,
		Signals: []OverlapSignal{
			{Key: "name_exact", Value: "docs skill", Score: 1},
		},
		Explanation: "Candidate name exactly matches an installed skill.",
		ExplanationMeta: ExplanationMetadata{
			Summary: "Exact collision on skill identity.",
			RuleIDs: []string{"OVLP.NAME.EXACT"},
			Signals: []OverlapSignal{
				{Key: "name_exact", Value: "docs skill", Score: 1},
			},
		},
	})
	report.Decision = &ConflictResolutionDecision{
		CandidateSkillID: "candidate.docs-skill",
		TargetSkillID:    "installed.docs-skill",
		Mode:             ResolutionUpdate,
		Blocking:         false,
		SelectedAt:       &selectedAt,
		Explanation:      "User chose to update the installed skill after reviewing the exact collision.",
		ExplanationMeta: ExplanationMetadata{
			Summary: "Exact collision can be handled as an update.",
			RuleIDs: []string{"OVLP.NAME.EXACT"},
			Signals: []OverlapSignal{
				{Key: "name_exact", Value: "docs skill", Score: 1},
			},
		},
	}

	got, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	want := `{"candidate":{"id":"candidate.docs-skill","name":"Docs Skill","description":"Generates a focused skill from docs.","in_scope":["Extract docs guidance","Render SKILL.md"],"out_of_scope":["Install files"],"commands":["cli-skill process --url https://example.com/docs"],"source_path":"/tmp/candidate/SKILL.md"},"overall_severity":"high","findings":[{"rule_id":"OVLP.NAME.EXACT","existing_skill_id":"installed.docs-skill","severity":"high","score":1,"signals":[{"key":"name_exact","value":"docs skill","score":1}],"explanation":"Candidate name exactly matches an installed skill.","explanation_meta":{"summary":"Exact collision on skill identity.","rule_ids":["OVLP.NAME.EXACT"],"signals":[{"key":"name_exact","value":"docs skill","score":1}]}}],"warnings":[{"path":"/Users/nick/.codex/skills/broken/SKILL.md","message":"parse failed"}],"decision":{"candidate_skill_id":"candidate.docs-skill","target_skill_id":"installed.docs-skill","mode":"update_existing","blocking":false,"selected_at":"2026-03-11T15:04:05Z","explanation":"User chose to update the installed skill after reviewing the exact collision.","explanation_meta":{"summary":"Exact collision can be handled as an update.","rule_ids":["OVLP.NAME.EXACT"],"signals":[{"key":"name_exact","value":"docs skill","score":1}]}}}`
	if string(got) != want {
		t.Fatalf("json contract mismatch\n got: %s\nwant: %s", string(got), want)
	}
}

func TestReportSortFindingsDeterministically(t *testing.T) {
	report := NewReport(SkillProfile{ID: "candidate"})
	report.AddFinding(OverlapFinding{
		RuleID:          "OVLP.SCOPE.SEMANTIC",
		ExistingSkillID: "skill-b",
		Severity:        SeverityMedium,
		Score:           0.7,
		Explanation:     "scope overlap",
	})
	report.AddFinding(OverlapFinding{
		RuleID:          "OVLP.NAME.EXACT",
		ExistingSkillID: "skill-a",
		Severity:        SeverityHigh,
		Score:           1,
		Explanation:     "name overlap",
	})
	report.AddFinding(OverlapFinding{
		RuleID:          "OVLP.CMD.OVERLAP",
		ExistingSkillID: "skill-c",
		Severity:        SeverityMedium,
		Score:           0.4,
		Explanation:     "command overlap",
	})

	gotRules := []string{
		report.Findings[0].RuleID,
		report.Findings[1].RuleID,
		report.Findings[2].RuleID,
	}
	wantRules := []string{"OVLP.NAME.EXACT", "OVLP.CMD.OVERLAP", "OVLP.SCOPE.SEMANTIC"}
	for i := range wantRules {
		if gotRules[i] != wantRules[i] {
			t.Fatalf("finding order[%d] = %q, want %q", i, gotRules[i], wantRules[i])
		}
	}
	if report.OverallSeverity != SeverityHigh {
		t.Fatalf("OverallSeverity = %q, want %q", report.OverallSeverity, SeverityHigh)
	}
}

func TestModelConflictResolutionDecisionResolutionState(t *testing.T) {
	selectedAt := time.Date(2026, time.March, 11, 17, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		decision ConflictResolutionDecision
		want     bool
	}{
		{
			name: "new install resolved",
			decision: ConflictResolutionDecision{
				Mode:       ResolutionNewInstall,
				Blocking:   false,
				SelectedAt: &selectedAt,
			},
			want: true,
		},
		{
			name: "blocking decision unresolved",
			decision: ConflictResolutionDecision{
				Mode:       ResolutionUpdate,
				Blocking:   true,
				SelectedAt: &selectedAt,
			},
			want: false,
		},
		{
			name: "abort never resolved for install handoff",
			decision: ConflictResolutionDecision{
				Mode:       ResolutionAbort,
				Blocking:   false,
				SelectedAt: &selectedAt,
			},
			want: false,
		},
		{
			name: "missing selection time unresolved",
			decision: ConflictResolutionDecision{
				Mode:     ResolutionMerge,
				Blocking: false,
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.decision.IsResolved(); got != tc.want {
				t.Fatalf("IsResolved() = %t, want %t", got, tc.want)
			}
		})
	}
}
