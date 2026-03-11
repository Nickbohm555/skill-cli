package install

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nickbohm555/skill-cli/internal/validation"
)

type TransactionResult struct {
	TargetDir string `json:"target_dir"`
	SkillPath string `json:"skill_path"`
}

type transactionFS interface {
	MkdirAll(path string, perm os.FileMode) error
	MkdirTemp(dir, pattern string) (string, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	ReadFile(name string) ([]byte, error)
	Rename(oldpath, newpath string) error
	RemoveAll(path string) error
	Stat(name string) (os.FileInfo, error)
}

type osTransactionFS struct{}

func (osTransactionFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osTransactionFS) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

func (osTransactionFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (osTransactionFS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (osTransactionFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (osTransactionFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (osTransactionFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

type TransactionExecutor struct {
	fs              transactionFS
	renderCandidate func(validation.CandidateSkill) string
	verifyStaged    func(transactionFS, string, validation.CandidateSkill) error
}

func InstallTransaction(request InstallRequest) (TransactionResult, error) {
	return NewTransactionExecutor().Install(request)
}

func NewTransactionExecutor() TransactionExecutor {
	return TransactionExecutor{
		fs:              osTransactionFS{},
		renderCandidate: RenderCandidateSkillMarkdown,
		verifyStaged:    verifyStagedInstall,
	}
}

func (e TransactionExecutor) Install(request InstallRequest) (TransactionResult, error) {
	if _, err := Preflight(request.ValidationReport, request.ConflictDecision); err != nil {
		return TransactionResult{}, err
	}
	if !request.Approval.IsExplicitApproval() {
		return TransactionResult{}, ErrInstallApprovalRequired
	}

	if e.fs == nil {
		e.fs = osTransactionFS{}
	}
	if e.renderCandidate == nil {
		e.renderCandidate = RenderCandidateSkillMarkdown
	}
	if e.verifyStaged == nil {
		e.verifyStaged = verifyStagedInstall
	}

	targetDir, err := resolveTransactionTargetDir(request.Target)
	if err != nil {
		return TransactionResult{}, err
	}

	parentDir := filepath.Dir(targetDir)
	if err := e.fs.MkdirAll(parentDir, 0o755); err != nil {
		return TransactionResult{}, fmt.Errorf("create install parent: %w", err)
	}

	stageDir, err := e.fs.MkdirTemp(parentDir, "."+filepath.Base(targetDir)+"-stage-*")
	if err != nil {
		return TransactionResult{}, fmt.Errorf("create stage dir: %w", err)
	}
	stageActive := true
	defer func() {
		if stageActive {
			_ = e.fs.RemoveAll(stageDir)
		}
	}()

	stageSkillPath := filepath.Join(stageDir, "SKILL.md")
	rendered := e.renderCandidate(request.Candidate.Skill)
	if err := e.fs.WriteFile(stageSkillPath, []byte(rendered), 0o644); err != nil {
		return TransactionResult{}, fmt.Errorf("write staged skill: %w", err)
	}

	if err := e.verifyStaged(e.fs, stageDir, request.Candidate.Skill); err != nil {
		return TransactionResult{}, err
	}

	existing, err := transactionPathExists(e.fs, targetDir)
	if err != nil {
		return TransactionResult{}, err
	}

	if !existing {
		if err := e.fs.Rename(stageDir, targetDir); err != nil {
			return TransactionResult{}, fmt.Errorf("activate staged install: %w", err)
		}
		stageActive = false
		return TransactionResult{
			TargetDir: targetDir,
			SkillPath: filepath.Join(targetDir, "SKILL.md"),
		}, nil
	}

	backupDir, err := reserveBackupPath(e.fs, parentDir, filepath.Base(targetDir))
	if err != nil {
		return TransactionResult{}, err
	}
	backupActive := false
	defer func() {
		if backupActive {
			_ = e.fs.RemoveAll(backupDir)
		}
	}()

	if err := e.fs.Rename(targetDir, backupDir); err != nil {
		return TransactionResult{}, fmt.Errorf("move existing install aside: %w", err)
	}
	backupActive = true

	if err := e.fs.Rename(stageDir, targetDir); err != nil {
		restoreErr := e.fs.Rename(backupDir, targetDir)
		if restoreErr != nil {
			return TransactionResult{}, errors.Join(
				fmt.Errorf("activate staged install: %w", err),
				fmt.Errorf("restore previous install: %w", restoreErr),
			)
		}
		backupActive = false
		if cleanupErr := e.fs.RemoveAll(stageDir); cleanupErr != nil {
			return TransactionResult{}, errors.Join(
				fmt.Errorf("activate staged install: %w", err),
				fmt.Errorf("cleanup failed stage dir: %w", cleanupErr),
			)
		}
		stageActive = false
		return TransactionResult{}, fmt.Errorf("activate staged install: %w", err)
	}

	stageActive = false
	if err := e.fs.RemoveAll(backupDir); err != nil {
		return TransactionResult{}, fmt.Errorf("cleanup backup install: %w", err)
	}
	backupActive = false

	return TransactionResult{
		TargetDir: targetDir,
		SkillPath: filepath.Join(targetDir, "SKILL.md"),
	}, nil
}

func resolveTransactionTargetDir(target InstallTarget) (string, error) {
	if skillDir := strings.TrimSpace(target.SkillDir); skillDir != "" {
		return skillDir, nil
	}
	if rootDir := strings.TrimSpace(target.RootDir); rootDir != "" && strings.TrimSpace(target.SkillID) != "" {
		return filepath.Join(rootDir, target.SkillID), nil
	}
	if existingPath := strings.TrimSpace(target.ExistingPath); existingPath != "" {
		return filepath.Dir(existingPath), nil
	}
	return "", fmt.Errorf("install target path is not configured")
}

func verifyStagedInstall(fs transactionFS, stageDir string, candidate validation.CandidateSkill) error {
	stageSkillPath := filepath.Join(stageDir, "SKILL.md")
	content, err := fs.ReadFile(stageSkillPath)
	if err != nil {
		return fmt.Errorf("verify staged skill read: %w", err)
	}

	parsed, err := validation.ParseSkill(content)
	if err != nil {
		return fmt.Errorf("verify staged skill parse: %w", err)
	}

	if parsed.Metadata.Name != candidate.Metadata.Name {
		return fmt.Errorf(
			"verify staged skill metadata.name mismatch: got %q want %q",
			parsed.Metadata.Name,
			candidate.Metadata.Name,
		)
	}
	if parsed.Title != candidate.Title {
		return fmt.Errorf("verify staged skill title mismatch: got %q want %q", parsed.Title, candidate.Title)
	}

	return nil
}

func transactionPathExists(fs transactionFS, path string) (bool, error) {
	_, err := fs.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat install target: %w", err)
}

func reserveBackupPath(fs transactionFS, parentDir, baseName string) (string, error) {
	placeholder, err := fs.MkdirTemp(parentDir, "."+baseName+"-backup-*")
	if err != nil {
		return "", fmt.Errorf("create backup placeholder: %w", err)
	}
	if err := fs.RemoveAll(placeholder); err != nil {
		return "", fmt.Errorf("clear backup placeholder: %w", err)
	}
	return placeholder, nil
}
