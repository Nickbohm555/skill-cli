package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nickbohm555/skill-cli/internal/validation"
)

const restartCodexGuidance = "If Codex does not show the installed skill yet, restart Codex and try again."

type ActivationVerificationResult struct {
	TargetDir           string                      `json:"target_dir"`
	SkillPath           string                      `json:"skill_path"`
	SkillID             string                      `json:"skill_id,omitempty"`
	Present             bool                        `json:"present"`
	Parsed              bool                        `json:"parsed"`
	Valid               bool                        `json:"valid"`
	Discoverable        bool                        `json:"discoverable"`
	ReadyNow            bool                        `json:"ready_now"`
	RestartFallback     bool                        `json:"restart_fallback"`
	VerificationMessage string                      `json:"verification_message,omitempty"`
	FallbackGuidance    string                      `json:"fallback_guidance,omitempty"`
	BlockingIssue       *validation.ValidationIssue `json:"blocking_issue,omitempty"`
}

type ActivationProbe struct {
	Request        InstallRequest
	Transaction    TransactionResult
	InstalledSkill validation.CandidateSkill
}

type DiscoverabilitySignal struct {
	Discoverable bool   `json:"discoverable"`
	Signal       string `json:"signal,omitempty"`
	Message      string `json:"message,omitempty"`
}

type activationFS interface {
	ReadFile(name string) ([]byte, error)
	Stat(name string) (os.FileInfo, error)
}

type osActivationFS struct{}

func (osActivationFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (osActivationFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

type ActivationVerifier struct {
	fs            activationFS
	parseSkill    func([]byte) (validation.CandidateSkill, error)
	validateSkill func(validation.CandidateSkill) validation.ValidationReport
	discoverable  func(ActivationProbe) (DiscoverabilitySignal, error)
}

func NewActivationVerifier() ActivationVerifier {
	return ActivationVerifier{
		fs:            osActivationFS{},
		parseSkill:    validation.ParseSkill,
		validateSkill: validateInstalledSkill,
		discoverable:  defaultDiscoverabilitySignal,
	}
}

func VerifyInstalledSkill(request InstallRequest, transaction TransactionResult) (ActivationVerificationResult, error) {
	return NewActivationVerifier().Verify(request, transaction)
}

func (v ActivationVerifier) Verify(request InstallRequest, transaction TransactionResult) (ActivationVerificationResult, error) {
	if v.fs == nil {
		v.fs = osActivationFS{}
	}
	if v.parseSkill == nil {
		v.parseSkill = validation.ParseSkill
	}
	if v.validateSkill == nil {
		v.validateSkill = validateInstalledSkill
	}
	if v.discoverable == nil {
		v.discoverable = defaultDiscoverabilitySignal
	}

	result := ActivationVerificationResult{
		TargetDir: transaction.TargetDir,
		SkillPath: transaction.SkillPath,
		SkillID:   request.Target.SkillID,
	}
	if strings.TrimSpace(result.SkillPath) == "" {
		result.SkillPath = filepath.Join(result.TargetDir, "SKILL.md")
	}

	if err := verifyInstalledPath(v.fs, result.TargetDir); err != nil {
		return result, err
	}
	if err := verifyInstalledPath(v.fs, result.SkillPath); err != nil {
		return result, err
	}
	result.Present = true

	content, err := v.fs.ReadFile(result.SkillPath)
	if err != nil {
		return result, fmt.Errorf("verify installed skill read: %w", err)
	}

	parsed, err := v.parseSkill(content)
	if err != nil {
		return result, fmt.Errorf("verify installed skill parse: %w", err)
	}
	result.Parsed = true

	expectedName := strings.TrimSpace(request.Candidate.Skill.Metadata.Name)
	if expectedName != "" && parsed.Metadata.Name != expectedName {
		return result, fmt.Errorf(
			"verify installed skill metadata.name mismatch: got %q want %q",
			parsed.Metadata.Name,
			expectedName,
		)
	}

	report := v.validateSkill(parsed)
	if issue, ok := report.NextBlockingIssue(); ok {
		issueCopy := issue
		result.BlockingIssue = &issueCopy
		return result, fmt.Errorf("verify installed skill validation: %s (%s)", issue.Message, issue.RuleID)
	}
	result.Valid = true

	signal, err := v.discoverable(ActivationProbe{
		Request:        request,
		Transaction:    transaction,
		InstalledSkill: parsed,
	})
	if err != nil {
		return result, fmt.Errorf("verify installed skill discoverability: %w", err)
	}

	result.Discoverable = signal.Discoverable
	if signal.Discoverable {
		result.ReadyNow = true
		result.VerificationMessage = firstNonEmpty(signal.Message, "Installed skill is present, parse-valid, and ready to use now.")
		return result, nil
	}

	result.RestartFallback = true
	result.VerificationMessage = firstNonEmpty(signal.Message, "Installed skill is present and parse-valid, but an immediate discoverability signal is missing.")
	result.FallbackGuidance = restartCodexGuidance
	return result, nil
}

func defaultDiscoverabilitySignal(probe ActivationProbe) (DiscoverabilitySignal, error) {
	message := "Installed skill is present and parse-valid, and no discovery lag was detected."
	if skillID := strings.TrimSpace(probe.Request.Target.SkillID); skillID != "" {
		message = fmt.Sprintf("Installed skill %q is present and parse-valid, and no discovery lag was detected.", skillID)
	}

	return DiscoverabilitySignal{
		Discoverable: true,
		Signal:       "installed_artifact_verified",
		Message:      message,
	}, nil
}

func verifyInstalledPath(fs activationFS, path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("verify installed skill path is empty")
	}

	if _, err := fs.Stat(path); err != nil {
		return fmt.Errorf("verify installed skill path %q: %w", path, err)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func validateInstalledSkill(candidate validation.CandidateSkill) validation.ValidationReport {
	report := validation.NewReport()
	structural := validation.ValidateStructural(candidate)
	report.AddIssues(structural.Issues...)
	semantic := validation.ValidateSemantic(candidate)
	report.AddIssues(semantic.Issues...)
	return report
}
