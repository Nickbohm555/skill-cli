package install

import (
	"time"

	"github.com/Nickbohm555/skill-cli/internal/overlap"
	"github.com/Nickbohm555/skill-cli/internal/validation"
)

type ApprovalSource string

const (
	ApprovalSourceNone               ApprovalSource = "none"
	ApprovalSourceInteractiveConfirm ApprovalSource = "interactive_confirm"
	ApprovalSourceNonInteractiveFlag ApprovalSource = "non_interactive_flag"
	ApprovalSourceDeclined           ApprovalSource = "declined"
	ApprovalSourceInterrupted        ApprovalSource = "interrupted"
)

type InstallCandidate struct {
	Skill      validation.CandidateSkill `json:"skill"`
	SourcePath string                    `json:"source_path,omitempty"`
	SkillID    string                    `json:"skill_id,omitempty"`
}

type InstallTarget struct {
	RootDir      string `json:"root_dir,omitempty"`
	SkillDir     string `json:"skill_dir,omitempty"`
	SkillID      string `json:"skill_id,omitempty"`
	ExistingPath string `json:"existing_path,omitempty"`
}

type ApprovalDecision struct {
	Approved       bool           `json:"approved"`
	ApprovalSource ApprovalSource `json:"approval_source"`
	DecisionAt     *time.Time     `json:"decision_at,omitempty"`
	Explanation    string         `json:"explanation,omitempty"`
}

type PreflightBlockReason string

const (
	PreflightBlockReasonNone               PreflightBlockReason = "none"
	PreflightBlockReasonValidationBlocking PreflightBlockReason = "validation_blocking"
	PreflightBlockReasonConflictMissing    PreflightBlockReason = "conflict_missing"
	PreflightBlockReasonConflictUnresolved PreflightBlockReason = "conflict_unresolved"
	PreflightBlockReasonConflictAbort      PreflightBlockReason = "conflict_abort"
)

type PreflightStatus struct {
	Allowed                 bool                                `json:"allowed"`
	Reason                  PreflightBlockReason                `json:"reason"`
	ErrorCode               ErrorCode                           `json:"error_code,omitempty"`
	Message                 string                              `json:"message,omitempty"`
	BlockingValidationIssue *validation.ValidationIssue         `json:"blocking_validation_issue,omitempty"`
	ConflictDecision        *overlap.ConflictResolutionDecision `json:"conflict_decision,omitempty"`
}

func (d ApprovalDecision) IsExplicitApproval() bool {
	if !d.Approved || d.DecisionAt == nil {
		return false
	}

	switch d.ApprovalSource {
	case ApprovalSourceInteractiveConfirm, ApprovalSourceNonInteractiveFlag:
		return true
	default:
		return false
	}
}

func (d ApprovalDecision) IsDenied() bool {
	return !d.Approved
}

type InstallRequest struct {
	Candidate        InstallCandidate                    `json:"candidate"`
	Target           InstallTarget                       `json:"target"`
	ValidationReport validation.ValidationReport         `json:"validation_report"`
	ConflictDecision *overlap.ConflictResolutionDecision `json:"conflict_decision,omitempty"`
	Approval         ApprovalDecision                    `json:"approval"`
	Interactive      bool                                `json:"interactive"`
}

func (r InstallRequest) ValidationResolved() bool {
	return !r.ValidationReport.HasBlockingIssues()
}

func (r InstallRequest) ConflictResolved() bool {
	if r.ConflictDecision == nil {
		return false
	}
	return r.ConflictDecision.IsResolved()
}

func (r InstallRequest) ReadyForWrite() bool {
	return r.ValidationResolved() && r.ConflictResolved() && r.Approval.IsExplicitApproval()
}

type InstallResult struct {
	Candidate        InstallCandidate                    `json:"candidate"`
	Target           InstallTarget                       `json:"target"`
	ValidationReport validation.ValidationReport         `json:"validation_report"`
	ConflictDecision *overlap.ConflictResolutionDecision `json:"conflict_decision,omitempty"`
	Approval         ApprovalDecision                    `json:"approval"`
	Preflight        PreflightStatus                     `json:"preflight"`
	PreflightPassed  bool                                `json:"preflight_passed"`
	WriteReady       bool                                `json:"write_ready"`
	Installed        bool                                `json:"installed"`
}

func NewInstallResult(request InstallRequest) InstallResult {
	preflight := PreflightStatus{
		Allowed:          request.ValidationResolved() && request.ConflictResolved(),
		Reason:           PreflightBlockReasonNone,
		ConflictDecision: request.ConflictDecision,
	}

	if issue, ok := request.ValidationReport.NextBlockingIssue(); ok {
		preflight.Allowed = false
		preflight.Reason = PreflightBlockReasonValidationBlocking
		preflight.ErrorCode = ErrorBlockedValidation
		preflight.Message = ErrInstallBlockedValidation.Error()
		preflight.BlockingValidationIssue = &issue
	} else if request.ConflictDecision == nil {
		preflight.Allowed = false
		preflight.Reason = PreflightBlockReasonConflictMissing
		preflight.ErrorCode = ErrorBlockedConflict
		preflight.Message = ErrInstallBlockedConflict.Error()
	} else if request.ConflictDecision.Mode == overlap.ResolutionAbort {
		preflight.Allowed = false
		preflight.Reason = PreflightBlockReasonConflictAbort
		preflight.ErrorCode = ErrorBlockedConflict
		preflight.Message = ErrInstallBlockedConflict.Error()
	} else if !request.ConflictDecision.IsResolved() {
		preflight.Allowed = false
		preflight.Reason = PreflightBlockReasonConflictUnresolved
		preflight.ErrorCode = ErrorBlockedConflict
		preflight.Message = ErrInstallBlockedConflict.Error()
	}

	return InstallResult{
		Candidate:        request.Candidate,
		Target:           request.Target,
		ValidationReport: request.ValidationReport,
		ConflictDecision: request.ConflictDecision,
		Approval:         request.Approval,
		Preflight:        preflight,
		PreflightPassed:  preflight.Allowed,
		WriteReady:       preflight.Allowed && request.Approval.IsExplicitApproval(),
		Installed:        false,
	}
}
