package install

import (
	"github.com/Nickbohm555/skill-cli/internal/overlap"
	"github.com/Nickbohm555/skill-cli/internal/validation"
)

func Preflight(report validation.ValidationReport, decision *overlap.ConflictResolutionDecision) (PreflightStatus, error) {
	if issue, ok := report.NextBlockingIssue(); ok {
		return PreflightStatus{
			Allowed:                 false,
			Reason:                  PreflightBlockReasonValidationBlocking,
			ErrorCode:               ErrorBlockedValidation,
			Message:                 ErrInstallBlockedValidation.Error(),
			BlockingValidationIssue: &issue,
			ConflictDecision:        decision,
		}, ErrInstallBlockedValidation
	}

	if decision == nil {
		return PreflightStatus{
			Allowed:   false,
			Reason:    PreflightBlockReasonConflictMissing,
			ErrorCode: ErrorBlockedConflict,
			Message:   ErrInstallBlockedConflict.Error(),
		}, ErrInstallBlockedConflict
	}

	if decision.Mode == overlap.ResolutionAbort {
		return PreflightStatus{
			Allowed:          false,
			Reason:           PreflightBlockReasonConflictAbort,
			ErrorCode:        ErrorBlockedConflict,
			Message:          ErrInstallBlockedConflict.Error(),
			ConflictDecision: decision,
		}, ErrInstallBlockedConflict
	}

	if !decision.IsResolved() {
		return PreflightStatus{
			Allowed:          false,
			Reason:           PreflightBlockReasonConflictUnresolved,
			ErrorCode:        ErrorBlockedConflict,
			Message:          ErrInstallBlockedConflict.Error(),
			ConflictDecision: decision,
		}, ErrInstallBlockedConflict
	}

	return PreflightStatus{
		Allowed:          true,
		Reason:           PreflightBlockReasonNone,
		ConflictDecision: decision,
	}, nil
}
