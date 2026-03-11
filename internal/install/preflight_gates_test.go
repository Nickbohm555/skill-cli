package install

import (
	"errors"
	"testing"
	"time"

	"github.com/Nickbohm555/skill-cli/internal/overlap"
	"github.com/Nickbohm555/skill-cli/internal/validation"
)

func TestPreflightPassThrough(t *testing.T) {
	t.Parallel()

	selectedAt := time.Date(2026, time.March, 11, 22, 30, 0, 0, time.UTC)
	decision := &overlap.ConflictResolutionDecision{
		CandidateSkillID: "candidate.docs",
		TargetSkillID:    "installed.docs",
		Mode:             overlap.ResolutionUpdate,
		SelectedAt:       &selectedAt,
	}

	status, err := Preflight(validation.NewReport(), decision)
	if err != nil {
		t.Fatalf("Preflight() error = %v, want nil", err)
	}
	if !status.Allowed {
		t.Fatal("Preflight().Allowed = false, want true")
	}
	if status.Reason != PreflightBlockReasonNone {
		t.Fatalf("Preflight().Reason = %q, want %q", status.Reason, PreflightBlockReasonNone)
	}
	if status.ConflictDecision != decision {
		t.Fatal("Preflight().ConflictDecision did not preserve the resolved decision")
	}
}

func TestPreflightBlockedValidation(t *testing.T) {
	t.Parallel()

	report := validation.NewReport()
	report.AddIssue(validation.ValidationIssue{
		RuleID:   "VAL.STRUCT.METADATA_NAME_REQUIRED",
		Severity: validation.SeverityError,
		Path:     "metadata.name",
		Message:  "name missing",
		Priority: 10,
	})

	status, err := Preflight(report, nil)
	if !errors.Is(err, ErrInstallBlockedValidation) {
		t.Fatalf("Preflight() error = %v, want %v", err, ErrInstallBlockedValidation)
	}
	if status.Allowed {
		t.Fatal("Preflight().Allowed = true, want false")
	}
	if status.Reason != PreflightBlockReasonValidationBlocking {
		t.Fatalf("Preflight().Reason = %q, want %q", status.Reason, PreflightBlockReasonValidationBlocking)
	}
	if status.BlockingValidationIssue == nil {
		t.Fatal("Preflight().BlockingValidationIssue = nil, want first blocking issue")
	}
	if status.BlockingValidationIssue.RuleID != "VAL.STRUCT.METADATA_NAME_REQUIRED" {
		t.Fatalf("Preflight().BlockingValidationIssue.RuleID = %q, want validation rule", status.BlockingValidationIssue.RuleID)
	}
	if status.ErrorCode != ErrorBlockedValidation {
		t.Fatalf("Preflight().ErrorCode = %q, want %q", status.ErrorCode, ErrorBlockedValidation)
	}
}

func TestPreflightBlockedConflictMissingOrUnresolved(t *testing.T) {
	t.Parallel()

	selectedAt := time.Date(2026, time.March, 11, 22, 31, 0, 0, time.UTC)
	tests := []struct {
		name     string
		decision *overlap.ConflictResolutionDecision
		reason   PreflightBlockReason
	}{
		{
			name:   "missing conflict decision",
			reason: PreflightBlockReasonConflictMissing,
		},
		{
			name: "blocking conflict decision",
			decision: &overlap.ConflictResolutionDecision{
				CandidateSkillID: "candidate.docs",
				TargetSkillID:    "installed.docs",
				Mode:             overlap.ResolutionMerge,
				Blocking:         true,
				SelectedAt:       &selectedAt,
			},
			reason: PreflightBlockReasonConflictUnresolved,
		},
		{
			name: "conflict decision without timestamp",
			decision: &overlap.ConflictResolutionDecision{
				CandidateSkillID: "candidate.docs",
				TargetSkillID:    "installed.docs",
				Mode:             overlap.ResolutionMerge,
			},
			reason: PreflightBlockReasonConflictUnresolved,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, err := Preflight(validation.NewReport(), tc.decision)
			if !errors.Is(err, ErrInstallBlockedConflict) {
				t.Fatalf("Preflight() error = %v, want %v", err, ErrInstallBlockedConflict)
			}
			if status.Allowed {
				t.Fatal("Preflight().Allowed = true, want false")
			}
			if status.Reason != tc.reason {
				t.Fatalf("Preflight().Reason = %q, want %q", status.Reason, tc.reason)
			}
			if status.ErrorCode != ErrorBlockedConflict {
				t.Fatalf("Preflight().ErrorCode = %q, want %q", status.ErrorCode, ErrorBlockedConflict)
			}
		})
	}
}

func TestPreflightBlockedConflictAbort(t *testing.T) {
	t.Parallel()

	selectedAt := time.Date(2026, time.March, 11, 22, 32, 0, 0, time.UTC)
	decision := &overlap.ConflictResolutionDecision{
		CandidateSkillID: "candidate.docs",
		TargetSkillID:    "installed.docs",
		Mode:             overlap.ResolutionAbort,
		SelectedAt:       &selectedAt,
	}

	status, err := Preflight(validation.NewReport(), decision)
	if !errors.Is(err, ErrInstallBlockedConflict) {
		t.Fatalf("Preflight() error = %v, want %v", err, ErrInstallBlockedConflict)
	}
	if status.Allowed {
		t.Fatal("Preflight().Allowed = true, want false")
	}
	if status.Reason != PreflightBlockReasonConflictAbort {
		t.Fatalf("Preflight().Reason = %q, want %q", status.Reason, PreflightBlockReasonConflictAbort)
	}
	if status.ConflictDecision != decision {
		t.Fatal("Preflight().ConflictDecision did not preserve the abort decision")
	}
}
