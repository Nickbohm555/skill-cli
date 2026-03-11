package install

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Nickbohm555/skill-cli/internal/overlap"
	"github.com/Nickbohm555/skill-cli/internal/validation"
)

func TestModelInstallContractsJSON(t *testing.T) {
	t.Parallel()

	decisionAt := time.Date(2026, time.March, 11, 21, 15, 0, 0, time.UTC)
	selectedAt := time.Date(2026, time.March, 11, 21, 10, 0, 0, time.UTC)

	request := InstallRequest{
		Candidate: InstallCandidate{
			Skill: validation.CandidateSkill{
				Metadata: validation.SkillMetadata{
					Name:        "go-docs-skill",
					Description: "Generate a skill from one docs URL.",
				},
				Title:          "Go Docs Skill",
				PurposeSummary: validation.TextSection{Heading: "Purpose", Body: "Generate a scoped skill from one docs URL."},
			},
			SourcePath: "/tmp/generated/SKILL.md",
			SkillID:    "go-docs-skill",
		},
		Target: InstallTarget{
			RootDir:      "/Users/nick/.codex/skills",
			SkillDir:     "/Users/nick/.codex/skills/go-docs-skill",
			SkillID:      "go-docs-skill",
			ExistingPath: "/Users/nick/.codex/skills/go-docs-skill/SKILL.md",
		},
		ValidationReport: validation.NewReport(),
		ConflictDecision: &overlap.ConflictResolutionDecision{
			CandidateSkillID: "go-docs-skill",
			TargetSkillID:    "go-docs-skill",
			Mode:             overlap.ResolutionUpdate,
			Blocking:         false,
			SelectedAt:       &selectedAt,
			Explanation:      "User explicitly chose update after resolving overlap.",
		},
		Approval: ApprovalDecision{
			Approved:       true,
			ApprovalSource: ApprovalSourceInteractiveConfirm,
			DecisionAt:     &decisionAt,
			Explanation:    "User approved install after reviewing the preview.",
		},
		Interactive: true,
	}

	got, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	want := `{"candidate":{"skill":{"metadata":{"name":"go-docs-skill","description":"Generate a skill from one docs URL."},"title":"Go Docs Skill","purpose_summary":{"heading":"Purpose","body":"Generate a scoped skill from one docs URL."},"primary_tasks":{"heading":""},"success_criteria":{"heading":""},"constraints":{"heading":""},"dependencies":{"heading":""},"example_requests":{"heading":""},"example_outputs":{"heading":""},"in_scope":{"heading":""},"out_of_scope":{"heading":""}},"source_path":"/tmp/generated/SKILL.md","skill_id":"go-docs-skill"},"target":{"root_dir":"/Users/nick/.codex/skills","skill_dir":"/Users/nick/.codex/skills/go-docs-skill","skill_id":"go-docs-skill","existing_path":"/Users/nick/.codex/skills/go-docs-skill/SKILL.md"},"validation_report":{"issues":[]},"conflict_decision":{"candidate_skill_id":"go-docs-skill","target_skill_id":"go-docs-skill","mode":"update_existing","blocking":false,"selected_at":"2026-03-11T21:10:00Z","explanation":"User explicitly chose update after resolving overlap.","explanation_meta":{}},"approval":{"approved":true,"approval_source":"interactive_confirm","decision_at":"2026-03-11T21:15:00Z","explanation":"User approved install after reviewing the preview."},"interactive":true}`
	if string(got) != want {
		t.Fatalf("json contract mismatch\n got: %s\nwant: %s", string(got), want)
	}
}

func TestModelApprovalDecisionExplicitApproval(t *testing.T) {
	t.Parallel()

	decisionAt := time.Date(2026, time.March, 11, 22, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		decision ApprovalDecision
		want     bool
	}{
		{
			name: "interactive confirm approved",
			decision: ApprovalDecision{
				Approved:       true,
				ApprovalSource: ApprovalSourceInteractiveConfirm,
				DecisionAt:     &decisionAt,
			},
			want: true,
		},
		{
			name: "non interactive flag approved",
			decision: ApprovalDecision{
				Approved:       true,
				ApprovalSource: ApprovalSourceNonInteractiveFlag,
				DecisionAt:     &decisionAt,
			},
			want: true,
		},
		{
			name: "approved without timestamp is not explicit",
			decision: ApprovalDecision{
				Approved:       true,
				ApprovalSource: ApprovalSourceInteractiveConfirm,
			},
			want: false,
		},
		{
			name: "declined is never explicit",
			decision: ApprovalDecision{
				Approved:       false,
				ApprovalSource: ApprovalSourceDeclined,
				DecisionAt:     &decisionAt,
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.decision.IsExplicitApproval(); got != tc.want {
				t.Fatalf("IsExplicitApproval() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestModelInstallRequestReadyForWrite(t *testing.T) {
	t.Parallel()

	selectedAt := time.Date(2026, time.March, 11, 22, 10, 0, 0, time.UTC)
	approvedAt := time.Date(2026, time.March, 11, 22, 11, 0, 0, time.UTC)

	blockedReport := validation.NewReport()
	blockedReport.AddIssue(validation.ValidationIssue{
		RuleID:   "VAL.STRUCT.METADATA_NAME_REQUIRED",
		Severity: validation.SeverityError,
		Path:     "metadata.name",
		Message:  "name missing",
		Priority: 10,
	})

	baseDecision := &overlap.ConflictResolutionDecision{
		CandidateSkillID: "candidate",
		TargetSkillID:    "installed",
		Mode:             overlap.ResolutionMerge,
		Blocking:         false,
		SelectedAt:       &selectedAt,
	}

	baseApproval := ApprovalDecision{
		Approved:       true,
		ApprovalSource: ApprovalSourceInteractiveConfirm,
		DecisionAt:     &approvedAt,
	}

	tests := []struct {
		name    string
		request InstallRequest
		want    bool
	}{
		{
			name: "all prerequisites satisfied",
			request: InstallRequest{
				ValidationReport: validation.NewReport(),
				ConflictDecision: baseDecision,
				Approval:         baseApproval,
			},
			want: true,
		},
		{
			name: "blocking validation prevents write readiness",
			request: InstallRequest{
				ValidationReport: blockedReport,
				ConflictDecision: baseDecision,
				Approval:         baseApproval,
			},
			want: false,
		},
		{
			name: "missing conflict decision prevents write readiness",
			request: InstallRequest{
				ValidationReport: validation.NewReport(),
				Approval:         baseApproval,
			},
			want: false,
		},
		{
			name: "missing explicit approval prevents write readiness",
			request: InstallRequest{
				ValidationReport: validation.NewReport(),
				ConflictDecision: baseDecision,
				Approval: ApprovalDecision{
					Approved:       false,
					ApprovalSource: ApprovalSourceDeclined,
					DecisionAt:     &approvedAt,
				},
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.request.ReadyForWrite(); got != tc.want {
				t.Fatalf("ReadyForWrite() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestErrorsClassifyInstallErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		err   error
		code  ErrorCode
		check func(error) bool
	}{
		{
			name:  "blocked validation",
			err:   ErrInstallBlockedValidation,
			code:  ErrorBlockedValidation,
			check: IsBlockedValidation,
		},
		{
			name:  "blocked conflict",
			err:   ErrInstallBlockedConflict,
			code:  ErrorBlockedConflict,
			check: IsBlockedConflict,
		},
		{
			name:  "approval declined",
			err:   ErrInstallDeclined,
			code:  ErrorApprovalDeclined,
			check: IsApprovalDeclined,
		},
		{
			name:  "non interactive approval required",
			err:   ErrInstallApprovalRequiredNonInteractive,
			code:  ErrorApprovalRequiredNonInteractive,
			check: IsApprovalRequiredNonInteractive,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := errors.Join(errors.New("outer"), tc.err)
			if got := ErrorCodeOf(wrapped); got != tc.code {
				t.Fatalf("ErrorCodeOf() = %q, want %q", got, tc.code)
			}
			if !tc.check(wrapped) {
				t.Fatalf("classification helper returned false for %q", tc.code)
			}
		})
	}
}
