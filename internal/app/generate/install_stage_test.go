package generate

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Nickbohm555/skill-cli/internal/install"
	"github.com/Nickbohm555/skill-cli/internal/overlap"
	"github.com/Nickbohm555/skill-cli/internal/validation"
)

func TestInstallStageEnforcesStrictSequenceBeforeActivation(t *testing.T) {
	t.Parallel()

	request := installStageTestRequest()
	callOrder := make([]string, 0, 6)
	transaction := install.TransactionResult{
		TargetDir: request.Target.SkillDir,
		SkillPath: request.Target.ExistingPath,
	}
	activation := install.ActivationVerificationResult{
		TargetDir:           transaction.TargetDir,
		SkillPath:           transaction.SkillPath,
		Present:             true,
		Parsed:              true,
		Valid:               true,
		Discoverable:        true,
		ReadyNow:            true,
		VerificationMessage: "Installed skill is present, parse-valid, and ready to use now.",
	}

	stage := InstallStage{
		Preflight: func(report validation.ValidationReport, decision *overlap.ConflictResolutionDecision) (install.PreflightStatus, error) {
			callOrder = append(callOrder, "preflight")
			return install.PreflightStatus{Allowed: true, Reason: install.PreflightBlockReasonNone, ConflictDecision: decision}, nil
		},
		RenderPreview: func(request install.InstallRequest) string {
			callOrder = append(callOrder, "preview")
			return "preview"
		},
		LoadExistingSkill: func(target install.InstallTarget) (string, error) {
			callOrder = append(callOrder, "load-existing")
			return "---\nname: old\n---\n", nil
		},
		RenderDiff: func(request install.InstallRequest, existingSkill string) string {
			callOrder = append(callOrder, "diff")
			if existingSkill == "" {
				t.Fatal("RenderDiff existingSkill = empty, want loaded existing content")
			}
			return "diff"
		},
		RequestApproval: func(request install.InstallRequest, preview string, diff string) (install.ApprovalDecision, error) {
			callOrder = append(callOrder, "approval")
			if preview != "preview" || diff != "diff" {
				t.Fatalf("RequestApproval() got preview=%q diff=%q", preview, diff)
			}
			return request.Approval, nil
		},
		ExecuteTransaction: func(request install.InstallRequest) (install.TransactionResult, error) {
			callOrder = append(callOrder, "transaction")
			if !request.Approval.IsExplicitApproval() {
				t.Fatal("ExecuteTransaction() received non-explicit approval")
			}
			return transaction, nil
		},
		VerifyActivation: func(request install.InstallRequest, got install.TransactionResult) (install.ActivationVerificationResult, error) {
			callOrder = append(callOrder, "activation")
			if got != transaction {
				t.Fatalf("VerifyActivation() transaction = %#v, want %#v", got, transaction)
			}
			return activation, nil
		},
	}

	result, err := stage.Run(request)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	wantOrder := []string{"preflight", "preview", "load-existing", "diff", "approval", "transaction", "activation"}
	if !reflect.DeepEqual(callOrder, wantOrder) {
		t.Fatalf("call order = %#v, want %#v", callOrder, wantOrder)
	}
	if result.ApprovalSource != install.ApprovalSourceInteractiveConfirm {
		t.Fatalf("ApprovalSource = %q, want %q", result.ApprovalSource, install.ApprovalSourceInteractiveConfirm)
	}
	if result.InstallTarget != request.Target.SkillDir {
		t.Fatalf("InstallTarget = %q, want %q", result.InstallTarget, request.Target.SkillDir)
	}
	if !result.Installed {
		t.Fatal("Installed = false, want true")
	}
	if !result.ReadyNow {
		t.Fatal("ReadyNow = false, want true")
	}
	if result.Activation == nil || !result.Activation.ReadyNow {
		t.Fatalf("Activation = %#v, want ready-now activation result", result.Activation)
	}
}

func TestInstallStageBlocksOnPreflightBeforePreviewOrWrite(t *testing.T) {
	t.Parallel()

	request := installStageTestRequest()
	blockedReport := validation.NewReport()
	blockedReport.AddIssue(validation.ValidationIssue{
		RuleID:   "VAL.SEMANTIC.BOUNDARY_REQUIRED",
		Severity: validation.SeverityError,
		Path:     "constraints",
		Message:  "boundary missing",
		Priority: 10,
	})
	request.ValidationReport = blockedReport

	callOrder := make([]string, 0, 1)
	stage := InstallStage{
		Preflight: func(report validation.ValidationReport, decision *overlap.ConflictResolutionDecision) (install.PreflightStatus, error) {
			callOrder = append(callOrder, "preflight")
			issue, _ := report.NextBlockingIssue()
			return install.PreflightStatus{
				Allowed:                 false,
				Reason:                  install.PreflightBlockReasonValidationBlocking,
				ErrorCode:               install.ErrorBlockedValidation,
				Message:                 install.ErrInstallBlockedValidation.Error(),
				BlockingValidationIssue: &issue,
				ConflictDecision:        decision,
			}, install.ErrInstallBlockedValidation
		},
		RenderPreview: func(request install.InstallRequest) string {
			t.Fatal("RenderPreview() called after blocked preflight")
			return ""
		},
		LoadExistingSkill: func(target install.InstallTarget) (string, error) {
			t.Fatal("LoadExistingSkill() called after blocked preflight")
			return "", nil
		},
		RenderDiff: func(request install.InstallRequest, existingSkill string) string {
			t.Fatal("RenderDiff() called after blocked preflight")
			return ""
		},
		RequestApproval: func(request install.InstallRequest, preview string, diff string) (install.ApprovalDecision, error) {
			t.Fatal("RequestApproval() called after blocked preflight")
			return install.ApprovalDecision{}, nil
		},
		ExecuteTransaction: func(request install.InstallRequest) (install.TransactionResult, error) {
			t.Fatal("ExecuteTransaction() called after blocked preflight")
			return install.TransactionResult{}, nil
		},
		VerifyActivation: func(request install.InstallRequest, transaction install.TransactionResult) (install.ActivationVerificationResult, error) {
			t.Fatal("VerifyActivation() called after blocked preflight")
			return install.ActivationVerificationResult{}, nil
		},
	}

	result, err := stage.Run(request)
	if !errors.Is(err, install.ErrInstallBlockedValidation) {
		t.Fatalf("Run() error = %v, want %v", err, install.ErrInstallBlockedValidation)
	}
	if !reflect.DeepEqual(callOrder, []string{"preflight"}) {
		t.Fatalf("call order = %#v, want only preflight", callOrder)
	}
	if result.Installed {
		t.Fatal("Installed = true, want false")
	}
	if result.Preflight.Reason != install.PreflightBlockReasonValidationBlocking {
		t.Fatalf("Preflight.Reason = %q, want %q", result.Preflight.Reason, install.PreflightBlockReasonValidationBlocking)
	}
}

func TestInstallStageDeclinedApprovalShortCircuitsBeforeTransaction(t *testing.T) {
	t.Parallel()

	request := installStageTestRequest()
	declinedAt := time.Date(2026, time.March, 11, 23, 45, 0, 0, time.UTC)

	callOrder := make([]string, 0, 5)
	stage := InstallStage{
		Preflight: func(report validation.ValidationReport, decision *overlap.ConflictResolutionDecision) (install.PreflightStatus, error) {
			callOrder = append(callOrder, "preflight")
			return install.PreflightStatus{Allowed: true, Reason: install.PreflightBlockReasonNone, ConflictDecision: decision}, nil
		},
		RenderPreview: func(request install.InstallRequest) string {
			callOrder = append(callOrder, "preview")
			return "preview"
		},
		LoadExistingSkill: func(target install.InstallTarget) (string, error) {
			callOrder = append(callOrder, "load-existing")
			return "", nil
		},
		RenderDiff: func(request install.InstallRequest, existingSkill string) string {
			callOrder = append(callOrder, "diff")
			return "diff"
		},
		RequestApproval: func(request install.InstallRequest, preview string, diff string) (install.ApprovalDecision, error) {
			callOrder = append(callOrder, "approval")
			return install.ApprovalDecision{
				Approved:       false,
				ApprovalSource: install.ApprovalSourceDeclined,
				DecisionAt:     &declinedAt,
				Explanation:    "User explicitly declined install approval.",
			}, install.ErrInstallDeclined
		},
		ExecuteTransaction: func(request install.InstallRequest) (install.TransactionResult, error) {
			t.Fatal("ExecuteTransaction() called after approval decline")
			return install.TransactionResult{}, nil
		},
		VerifyActivation: func(request install.InstallRequest, transaction install.TransactionResult) (install.ActivationVerificationResult, error) {
			t.Fatal("VerifyActivation() called after approval decline")
			return install.ActivationVerificationResult{}, nil
		},
	}

	result, err := stage.Run(request)
	if !errors.Is(err, install.ErrInstallDeclined) {
		t.Fatalf("Run() error = %v, want %v", err, install.ErrInstallDeclined)
	}

	wantOrder := []string{"preflight", "preview", "load-existing", "diff", "approval"}
	if !reflect.DeepEqual(callOrder, wantOrder) {
		t.Fatalf("call order = %#v, want %#v", callOrder, wantOrder)
	}
	if result.Installed {
		t.Fatal("Installed = true, want false")
	}
	if result.ApprovalSource != install.ApprovalSourceDeclined {
		t.Fatalf("ApprovalSource = %q, want %q", result.ApprovalSource, install.ApprovalSourceDeclined)
	}
	if result.Transaction != nil {
		t.Fatalf("Transaction = %#v, want nil", result.Transaction)
	}
	if result.Activation != nil {
		t.Fatalf("Activation = %#v, want nil", result.Activation)
	}
}

func installStageTestRequest() install.InstallRequest {
	selectedAt := time.Date(2026, time.March, 11, 21, 10, 0, 0, time.UTC)
	approvedAt := time.Date(2026, time.March, 11, 21, 20, 0, 0, time.UTC)

	return install.InstallRequest{
		Candidate: install.InstallCandidate{
			SkillID:    "go-docs-skill",
			SourcePath: "/tmp/generated/SKILL.md",
			Skill: validation.CandidateSkill{
				Metadata: validation.SkillMetadata{
					Name:        "go-docs-skill",
					Description: "Generate a skill from one docs URL.",
				},
				Title:          "Go Docs Skill",
				PurposeSummary: validation.TextSection{Heading: "Purpose", Body: "Generate a scoped skill from one docs URL."},
			},
		},
		Target: install.InstallTarget{
			RootDir:      "/tmp/codex/skills",
			SkillDir:     "/tmp/codex/skills/go-docs-skill",
			SkillID:      "go-docs-skill",
			ExistingPath: "/tmp/codex/skills/go-docs-skill/SKILL.md",
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
		Approval: install.ApprovalDecision{
			Approved:       true,
			ApprovalSource: install.ApprovalSourceInteractiveConfirm,
			DecisionAt:     &approvedAt,
			Explanation:    "User explicitly approved install after preview.",
		},
		Interactive: true,
	}
}
