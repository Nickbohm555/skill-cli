package install

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Nickbohm555/skill-cli/internal/overlap"
)

func TestTransactionCreateWritesApprovedSkill(t *testing.T) {
	rootDir := t.TempDir()
	request := transactionTestRequest(rootDir)

	result, err := InstallTransaction(request)
	if err != nil {
		t.Fatalf("InstallTransaction() error = %v, want nil", err)
	}

	wantTargetDir := filepath.Join(rootDir, "go-docs-skill")
	if result.TargetDir != wantTargetDir {
		t.Fatalf("InstallTransaction().TargetDir = %q, want %q", result.TargetDir, wantTargetDir)
	}

	got, err := os.ReadFile(filepath.Join(wantTargetDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(installed skill) error = %v", err)
	}

	want := RenderCandidateSkillMarkdown(request.Candidate.Skill)
	if string(got) != want {
		t.Fatalf("installed SKILL.md mismatch\n got:\n%s\nwant:\n%s", string(got), want)
	}

	assertNoTransactionArtifacts(t, rootDir)
}

func TestTransactionUpdateReplacesExistingSkillAtomically(t *testing.T) {
	rootDir := t.TempDir()
	request := transactionTestRequest(rootDir)
	targetDir := filepath.Join(rootDir, "go-docs-skill")

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(existing target) error = %v", err)
	}
	oldContent := "---\nname: go-docs-skill\ndescription: old\n---\n\n# Old Skill\n"
	if err := os.WriteFile(filepath.Join(targetDir, "SKILL.md"), []byte(oldContent), 0o644); err != nil {
		t.Fatalf("WriteFile(existing skill) error = %v", err)
	}

	result, err := InstallTransaction(request)
	if err != nil {
		t.Fatalf("InstallTransaction() error = %v, want nil", err)
	}

	got, err := os.ReadFile(result.SkillPath)
	if err != nil {
		t.Fatalf("ReadFile(updated skill) error = %v", err)
	}

	want := RenderCandidateSkillMarkdown(request.Candidate.Skill)
	if string(got) != want {
		t.Fatalf("updated SKILL.md mismatch\n got:\n%s\nwant:\n%s", string(got), want)
	}

	assertNoTransactionArtifacts(t, rootDir)
}

func TestTransactionRollbackOnActivationFailureRestoresExistingSkill(t *testing.T) {
	rootDir := t.TempDir()
	request := transactionTestRequest(rootDir)
	targetDir := filepath.Join(rootDir, "go-docs-skill")

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(existing target) error = %v", err)
	}
	original := "---\nname: go-docs-skill\ndescription: old\n---\n\n# Old Skill\n"
	if err := os.WriteFile(filepath.Join(targetDir, "SKILL.md"), []byte(original), 0o644); err != nil {
		t.Fatalf("WriteFile(existing skill) error = %v", err)
	}

	executor := NewTransactionExecutor()
	executor.fs = renameFailFS{
		transactionFS: executor.fs,
		failTarget:    targetDir,
	}

	_, err := executor.Install(request)
	if err == nil {
		t.Fatal("Install() error = nil, want activation failure")
	}
	if !strings.Contains(err.Error(), "activate staged install") {
		t.Fatalf("Install() error = %v, want activation failure context", err)
	}

	got, readErr := os.ReadFile(filepath.Join(targetDir, "SKILL.md"))
	if readErr != nil {
		t.Fatalf("ReadFile(restored skill) error = %v", readErr)
	}
	if string(got) != original {
		t.Fatalf("restored SKILL.md mismatch\n got:\n%s\nwant:\n%s", string(got), original)
	}

	assertNoTransactionArtifacts(t, rootDir)
}

func TestTransactionBlocksMissingExplicitApproval(t *testing.T) {
	rootDir := t.TempDir()
	request := transactionTestRequest(rootDir)
	request.Approval = ApprovalDecision{}

	_, err := InstallTransaction(request)
	if !errors.Is(err, ErrInstallApprovalRequired) {
		t.Fatalf("InstallTransaction() error = %v, want %v", err, ErrInstallApprovalRequired)
	}

	targetDir := filepath.Join(rootDir, "go-docs-skill")
	if _, statErr := os.Stat(targetDir); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("Stat(targetDir) error = %v, want not exist", statErr)
	}
}

type renameFailFS struct {
	transactionFS
	failTarget string
}

func (fs renameFailFS) Rename(oldpath, newpath string) error {
	if newpath == fs.failTarget && strings.Contains(filepath.Base(oldpath), "-stage-") {
		return errors.New("injected rename failure")
	}
	return fs.transactionFS.Rename(oldpath, newpath)
}

func transactionTestRequest(rootDir string) InstallRequest {
	request := previewTestRequest()
	request.Target = InstallTarget{
		RootDir:      rootDir,
		SkillDir:     filepath.Join(rootDir, "go-docs-skill"),
		SkillID:      "go-docs-skill",
		ExistingPath: filepath.Join(rootDir, "go-docs-skill", "SKILL.md"),
	}

	selectedAt := time.Date(2026, time.March, 11, 21, 10, 0, 0, time.UTC)
	approvedAt := time.Date(2026, time.March, 11, 21, 20, 0, 0, time.UTC)
	request.ConflictDecision = &overlap.ConflictResolutionDecision{
		CandidateSkillID: "go-docs-skill",
		TargetSkillID:    "go-docs-skill",
		Mode:             overlap.ResolutionUpdate,
		Blocking:         false,
		SelectedAt:       &selectedAt,
	}
	request.Approval = ApprovalDecision{
		Approved:       true,
		ApprovalSource: ApprovalSourceInteractiveConfirm,
		DecisionAt:     &approvedAt,
		Explanation:    "User explicitly approved install after preview.",
	}

	return request
}

func assertNoTransactionArtifacts(t *testing.T, rootDir string) {
	t.Helper()

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		t.Fatalf("ReadDir(rootDir) error = %v", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, "-stage-") || strings.Contains(name, "-backup-") {
			t.Fatalf("unexpected transaction artifact left behind: %s", name)
		}
	}
}
