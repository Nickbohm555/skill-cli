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
	PreflightPassed  bool                                `json:"preflight_passed"`
	WriteReady       bool                                `json:"write_ready"`
	Installed        bool                                `json:"installed"`
}

func NewInstallResult(request InstallRequest) InstallResult {
	preflightPassed := request.ValidationResolved() && request.ConflictResolved()
	return InstallResult{
		Candidate:        request.Candidate,
		Target:           request.Target,
		ValidationReport: request.ValidationReport,
		ConflictDecision: request.ConflictDecision,
		Approval:         request.Approval,
		PreflightPassed:  preflightPassed,
		WriteReady:       preflightPassed && request.Approval.IsExplicitApproval(),
		Installed:        false,
	}
}
