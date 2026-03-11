package install

import (
	"os"
	"strings"
	"testing"
)

func TestActivateINST04ImmediateUsabilityAndExceptionalFallback(t *testing.T) {
	t.Parallel()

	t.Run("ready now is default success path", func(t *testing.T) {
		rootDir := t.TempDir()
		request := activationTestRequest(rootDir)

		transaction, err := InstallTransaction(request)
		if err != nil {
			t.Fatalf("InstallTransaction() error = %v, want nil", err)
		}

		result, err := VerifyInstalledSkill(request, transaction)
		if err != nil {
			t.Fatalf("VerifyInstalledSkill() error = %v, want nil", err)
		}
		if !result.ReadyNow {
			t.Fatal("ReadyNow = false, want true")
		}
		if result.RestartFallback {
			t.Fatal("RestartFallback = true, want false for normal success path")
		}
		if result.FallbackGuidance != "" {
			t.Fatalf("FallbackGuidance = %q, want empty on normal success path", result.FallbackGuidance)
		}
	})

	t.Run("restart guidance appears only when discoverability signal is missing", func(t *testing.T) {
		rootDir := t.TempDir()
		request := activationTestRequest(rootDir)

		transaction, err := InstallTransaction(request)
		if err != nil {
			t.Fatalf("InstallTransaction() error = %v, want nil", err)
		}

		verifier := NewActivationVerifier()
		verifier.discoverable = func(probe ActivationProbe) (DiscoverabilitySignal, error) {
			return DiscoverabilitySignal{
				Discoverable: false,
				Signal:       "codex_registry_missing",
				Message:      "Installed skill parsed correctly, but the discoverability signal is not visible yet.",
			}, nil
		}

		result, err := verifier.Verify(request, transaction)
		if err != nil {
			t.Fatalf("Verify() error = %v, want nil", err)
		}
		if result.ReadyNow {
			t.Fatal("ReadyNow = true, want false when discoverability signal is missing")
		}
		if !result.RestartFallback {
			t.Fatal("RestartFallback = false, want true when discoverability signal is missing")
		}
		if result.FallbackGuidance != restartCodexGuidance {
			t.Fatalf("FallbackGuidance = %q, want %q", result.FallbackGuidance, restartCodexGuidance)
		}
	})
}

func TestActivateReadyNowSuccess(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	request := activationTestRequest(rootDir)

	transaction, err := InstallTransaction(request)
	if err != nil {
		t.Fatalf("InstallTransaction() error = %v, want nil", err)
	}

	result, err := VerifyInstalledSkill(request, transaction)
	if err != nil {
		t.Fatalf("VerifyInstalledSkill() error = %v, want nil", err)
	}
	if !result.Present {
		t.Fatal("VerifyInstalledSkill().Present = false, want true")
	}
	if !result.Parsed {
		t.Fatal("VerifyInstalledSkill().Parsed = false, want true")
	}
	if !result.Valid {
		t.Fatal("VerifyInstalledSkill().Valid = false, want true")
	}
	if !result.Discoverable {
		t.Fatal("VerifyInstalledSkill().Discoverable = false, want true")
	}
	if !result.ReadyNow {
		t.Fatal("VerifyInstalledSkill().ReadyNow = false, want true")
	}
	if result.RestartFallback {
		t.Fatal("VerifyInstalledSkill().RestartFallback = true, want false")
	}
	if result.FallbackGuidance != "" {
		t.Fatalf("VerifyInstalledSkill().FallbackGuidance = %q, want empty", result.FallbackGuidance)
	}
	if !strings.Contains(result.VerificationMessage, "parse-valid") {
		t.Fatalf("VerifyInstalledSkill().VerificationMessage = %q, want parse-valid success message", result.VerificationMessage)
	}
}

func TestActivateParseInvalidFailsClosed(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	request := activationTestRequest(rootDir)
	transaction := TransactionResult{
		TargetDir: request.Target.SkillDir,
		SkillPath: request.Target.ExistingPath,
	}

	if err := os.MkdirAll(transaction.TargetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(transaction.SkillPath, []byte("---\nname: [\n---\n\n# Broken Skill\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	result, err := VerifyInstalledSkill(request, transaction)
	if err == nil {
		t.Fatal("VerifyInstalledSkill() error = nil, want parse failure")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("VerifyInstalledSkill() error = %v, want parse context", err)
	}
	if !result.Present {
		t.Fatal("VerifyInstalledSkill().Present = false, want true once files exist")
	}
	if result.Parsed {
		t.Fatal("VerifyInstalledSkill().Parsed = true, want false")
	}
	if result.Valid {
		t.Fatal("VerifyInstalledSkill().Valid = true, want false")
	}
}

func TestActivateFallbackGuidanceOnlyWhenDiscoverabilitySignalMissing(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	request := activationTestRequest(rootDir)

	transaction, err := InstallTransaction(request)
	if err != nil {
		t.Fatalf("InstallTransaction() error = %v, want nil", err)
	}

	verifier := NewActivationVerifier()
	verifier.discoverable = func(probe ActivationProbe) (DiscoverabilitySignal, error) {
		if probe.Transaction.TargetDir != transaction.TargetDir {
			t.Fatalf("discoverability probe target = %q, want %q", probe.Transaction.TargetDir, transaction.TargetDir)
		}
		return DiscoverabilitySignal{
			Discoverable: false,
			Signal:       "codex_registry_missing",
			Message:      "Installed skill parsed correctly, but the discoverability signal is not visible yet.",
		}, nil
	}

	result, err := verifier.Verify(request, transaction)
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
	if result.ReadyNow {
		t.Fatal("Verify().ReadyNow = true, want false")
	}
	if !result.RestartFallback {
		t.Fatal("Verify().RestartFallback = false, want true")
	}
	if result.Discoverable {
		t.Fatal("Verify().Discoverable = true, want false")
	}
	if result.FallbackGuidance != restartCodexGuidance {
		t.Fatalf("Verify().FallbackGuidance = %q, want %q", result.FallbackGuidance, restartCodexGuidance)
	}
	if !strings.Contains(result.VerificationMessage, "discoverability signal is not visible yet") {
		t.Fatalf("Verify().VerificationMessage = %q, want discoverability explanation", result.VerificationMessage)
	}
}

func activationTestRequest(rootDir string) InstallRequest {
	request := transactionTestRequest(rootDir)
	request.Candidate.Skill.Constraints.Heading = "Constraints"
	request.Candidate.Skill.Constraints.Items = []string{
		"Use only the provided docs URL as the source of truth.",
	}
	request.Candidate.Skill.Dependencies.Heading = "Dependencies"
	request.Candidate.Skill.Dependencies.Items = []string{
		"Go 1.25.x",
	}
	request.Candidate.Skill.ExampleRequests.Heading = "Example Requests"
	request.Candidate.Skill.ExampleRequests.Items = []string{
		"Generate a skill from https://go.dev/doc/.",
	}
	request.Candidate.Skill.ExampleOutputs.Heading = "Example Outputs"
	request.Candidate.Skill.ExampleOutputs.Items = []string{
		"A SKILL.md with install steps and source boundaries.",
	}
	return request
}
