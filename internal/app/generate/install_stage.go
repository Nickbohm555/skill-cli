package generate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nickbohm555/skill-cli/internal/install"
	"github.com/Nickbohm555/skill-cli/internal/overlap"
	"github.com/Nickbohm555/skill-cli/internal/validation"
)

type InstallPreflightFunc func(report validation.ValidationReport, decision *overlap.ConflictResolutionDecision) (install.PreflightStatus, error)

type InstallRenderPreviewFunc func(request install.InstallRequest) string

type InstallLoadExistingSkillFunc func(target install.InstallTarget) (string, error)

type InstallRenderDiffFunc func(request install.InstallRequest, existingSkill string) string

type InstallRequestApprovalFunc func(request install.InstallRequest, preview string, diff string) (install.ApprovalDecision, error)

type InstallExecuteTransactionFunc func(request install.InstallRequest) (install.TransactionResult, error)

type InstallVerifyActivationFunc func(request install.InstallRequest, transaction install.TransactionResult) (install.ActivationVerificationResult, error)

type InstallStage struct {
	Preflight          InstallPreflightFunc
	RenderPreview      InstallRenderPreviewFunc
	LoadExistingSkill  InstallLoadExistingSkillFunc
	RenderDiff         InstallRenderDiffFunc
	RequestApproval    InstallRequestApprovalFunc
	ExecuteTransaction InstallExecuteTransactionFunc
	VerifyActivation   InstallVerifyActivationFunc
}

type InstallStageResult struct {
	Request        install.InstallRequest                `json:"request"`
	Preflight      install.PreflightStatus               `json:"preflight"`
	Preview        string                                `json:"preview,omitempty"`
	Diff           string                                `json:"diff,omitempty"`
	Approval       install.ApprovalDecision              `json:"approval"`
	ApprovalSource install.ApprovalSource                `json:"approval_source"`
	InstallTarget  string                                `json:"install_target,omitempty"`
	Transaction    *install.TransactionResult            `json:"transaction,omitempty"`
	Activation     *install.ActivationVerificationResult `json:"activation,omitempty"`
	Installed      bool                                  `json:"installed"`
	ReadyNow       bool                                  `json:"ready_now"`
}

func NewInstallStage() InstallStage {
	return InstallStage{
		Preflight:          install.Preflight,
		RenderPreview:      install.RenderPreview,
		LoadExistingSkill:  loadExistingInstalledSkill,
		RenderDiff:         install.RenderDiff,
		RequestApproval:    defaultApprovalDecision,
		ExecuteTransaction: install.InstallTransaction,
		VerifyActivation:   install.VerifyInstalledSkill,
	}
}

func (s InstallStage) Run(request install.InstallRequest) (InstallStageResult, error) {
	if s.Preflight == nil {
		s.Preflight = install.Preflight
	}
	if s.RenderPreview == nil {
		s.RenderPreview = install.RenderPreview
	}
	if s.LoadExistingSkill == nil {
		s.LoadExistingSkill = loadExistingInstalledSkill
	}
	if s.RenderDiff == nil {
		s.RenderDiff = install.RenderDiff
	}
	if s.RequestApproval == nil {
		s.RequestApproval = defaultApprovalDecision
	}
	if s.ExecuteTransaction == nil {
		s.ExecuteTransaction = install.InstallTransaction
	}
	if s.VerifyActivation == nil {
		s.VerifyActivation = install.VerifyInstalledSkill
	}

	result := InstallStageResult{
		Request:        request,
		Approval:       request.Approval,
		ApprovalSource: request.Approval.ApprovalSource,
		InstallTarget:  resolveInstallTarget(request.Target),
	}

	preflight, err := s.Preflight(request.ValidationReport, request.ConflictDecision)
	result.Preflight = preflight
	if err != nil {
		return result, err
	}

	result.Preview = s.RenderPreview(request)

	existingSkill, err := s.LoadExistingSkill(request.Target)
	if err != nil {
		return result, err
	}
	result.Diff = s.RenderDiff(request, existingSkill)

	approval, err := s.RequestApproval(request, result.Preview, result.Diff)
	result.Approval = approval
	result.ApprovalSource = approval.ApprovalSource
	request.Approval = approval
	result.Request = request
	if err != nil {
		return result, err
	}

	transaction, err := s.ExecuteTransaction(request)
	if err != nil {
		return result, err
	}
	result.Transaction = &transaction
	result.InstallTarget = transaction.TargetDir
	result.Installed = true

	activation, err := s.VerifyActivation(request, transaction)
	result.Activation = &activation
	result.ReadyNow = activation.ReadyNow
	if err != nil {
		return result, err
	}

	return result, nil
}

func loadExistingInstalledSkill(target install.InstallTarget) (string, error) {
	existingPath := strings.TrimSpace(target.ExistingPath)
	if existingPath == "" {
		return "", nil
	}

	content, err := os.ReadFile(existingPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("load existing installed skill: %w", err)
	}

	return string(content), nil
}

func defaultApprovalDecision(request install.InstallRequest, preview string, diff string) (install.ApprovalDecision, error) {
	if request.Approval.IsExplicitApproval() {
		return request.Approval, nil
	}
	if hasRecordedApprovalDecision(request.Approval) && request.Approval.IsDenied() {
		return request.Approval, install.ErrInstallDeclined
	}
	return request.Approval, install.ErrInstallApprovalRequired
}

func hasRecordedApprovalDecision(decision install.ApprovalDecision) bool {
	return decision.DecisionAt != nil ||
		decision.ApprovalSource != install.ApprovalSourceNone ||
		strings.TrimSpace(decision.Explanation) != ""
}

func resolveInstallTarget(target install.InstallTarget) string {
	if skillDir := strings.TrimSpace(target.SkillDir); skillDir != "" {
		return skillDir
	}
	if rootDir := strings.TrimSpace(target.RootDir); rootDir != "" && strings.TrimSpace(target.SkillID) != "" {
		return filepath.Join(rootDir, target.SkillID)
	}
	if existingPath := strings.TrimSpace(target.ExistingPath); existingPath != "" {
		return filepath.Dir(existingPath)
	}
	return ""
}
