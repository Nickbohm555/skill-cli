package install

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Nickbohm555/skill-cli/internal/overlap"
	"github.com/Nickbohm555/skill-cli/internal/validation"
)

func TestPreviewCreateScenarioIsDeterministic(t *testing.T) {
	t.Parallel()

	request := previewTestRequest()
	request.ConflictDecision = &overlap.ConflictResolutionDecision{
		CandidateSkillID: "go-docs-skill",
		Mode:             overlap.ResolutionNewInstall,
	}
	request.Target.ExistingPath = ""

	first := RenderPreview(request)
	second := RenderPreview(request)
	if first != second {
		t.Fatalf("RenderPreview() output was not deterministic\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	wantContains := []string{
		"Install Preview",
		"Mode: new_install",
		"Target path: /Users/nick/.codex/skills/go-docs-skill",
		"Skill ID: go-docs-skill",
		"Metadata:",
		"- owner: docs-platform",
		"- source_url: https://go.dev/doc/",
		"## Purpose",
		"Generate a scoped Codex skill from one docs URL.",
		"## Primary Tasks",
		"- Fetch one docs root and normalize the extracted guidance.",
		"## In Scope",
		"- Converting one docs source into a reusable Codex skill.",
	}
	for _, want := range wantContains {
		if !strings.Contains(first, want) {
			t.Fatalf("RenderPreview() missing %q\noutput:\n%s", want, first)
		}
	}
}

func TestDiffUpdateScenarioIsDeterministic(t *testing.T) {
	t.Parallel()

	request := previewTestRequest()
	request.ConflictDecision = &overlap.ConflictResolutionDecision{
		CandidateSkillID: "go-docs-skill",
		TargetSkillID:    "go-docs-skill",
		Mode:             overlap.ResolutionUpdate,
	}

	existing := strings.TrimSpace(`---
name: go-docs-skill
description: Generate a skill from docs.
metadata:
  owner: docs-platform
  source_url: https://go.dev/old-doc/
---

# Go Docs Skill

## Purpose
Generate a skill from older docs.

## Primary Tasks
- Fetch one docs root and summarize it.

## Success Criteria
- Output remains installable.
`) + "\n"

	first := RenderDiff(request, existing)
	second := RenderDiff(request, existing)
	if first != second {
		t.Fatalf("RenderDiff() output was not deterministic\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	normalized := stripANSI(first)

	wantContains := []string{
		"Install Diff",
		"Mode: update_existing",
		"Target path: /Users/nick/.codex/skills/go-docs-skill",
		"source_url: https://go.dev/",
		"Generate a skill from",
		"Generate a scoped Codex skill from",
	}
	for _, want := range wantContains {
		if !strings.Contains(normalized, want) {
			t.Fatalf("RenderDiff() missing %q\noutput:\n%s", want, normalized)
		}
	}
}

func TestSequencePreviewAndDiffDoNotRequireApprovalOrWrite(t *testing.T) {
	t.Parallel()

	rootDir := t.TempDir()
	request := previewTestRequest()
	request.Target = InstallTarget{
		RootDir:      rootDir,
		SkillDir:     filepath.Join(rootDir, "go-docs-skill"),
		SkillID:      "go-docs-skill",
		ExistingPath: filepath.Join(rootDir, "go-docs-skill", "SKILL.md"),
	}
	request.Approval = ApprovalDecision{}

	preview := RenderPreview(request)
	diff := RenderDiff(request, "")

	if !strings.Contains(preview, "Install Preview") {
		t.Fatalf("RenderPreview() missing header\noutput:\n%s", preview)
	}
	if !strings.Contains(diff, "Install Diff") {
		t.Fatalf("RenderDiff() missing header\noutput:\n%s", diff)
	}
	if request.ReadyForWrite() {
		t.Fatal("request.ReadyForWrite() = true, want false before approval")
	}
	if _, err := os.Stat(request.Target.SkillDir); !os.IsNotExist(err) {
		t.Fatalf("Stat(target skill dir) error = %v, want not exist after preview-only calls", err)
	}
}

func previewTestRequest() InstallRequest {
	selectedAt := time.Date(2026, time.March, 11, 21, 10, 0, 0, time.UTC)

	return InstallRequest{
		Candidate: InstallCandidate{
			Skill: validation.CandidateSkill{
				Metadata: validation.SkillMetadata{
					Name:        "go-docs-skill",
					Description: "Generate a skill from one docs URL.",
					Extra: map[string]string{
						"source_url": "https://go.dev/doc/",
						"owner":      "docs-platform",
					},
				},
				Title: "Go Docs Skill",
				PurposeSummary: validation.TextSection{
					Heading: "Purpose",
					Body:    "Generate a scoped Codex skill from one docs URL.",
				},
				PrimaryTasks: validation.ListSection{
					Heading: "Primary Tasks",
					Items: []string{
						"Fetch one docs root and normalize the extracted guidance.",
						"Render install-ready skill content with explicit boundaries.",
					},
				},
				SuccessCriteria: validation.ListSection{
					Heading: "Success Criteria",
					Items: []string{
						"The generated skill is installable and constrained to one docs source.",
					},
				},
				InScope: validation.ListSection{
					Heading: "In Scope",
					Items: []string{
						"Converting one docs source into a reusable Codex skill.",
					},
				},
				OutOfScope: validation.ListSection{
					Heading: "Out Of Scope",
					Items: []string{
						"Editing installed skills by hand after generation.",
					},
				},
			},
			SourcePath: "/tmp/generated/SKILL.md",
			SkillID:    "go-docs-skill",
		},
		Target: InstallTarget{
			RootDir:      "/Users/nick/.codex/skills",
			SkillDir:     "/Users/nick/.codex/skills/go-docs-skill",
			SkillID:      "go-docs-skill",
			ExistingPath: "/Users/nick/.codex/skills/go-docs-skill/SKILL.md",
		},
		ValidationReport: validation.NewReport(),
		ConflictDecision: &overlap.ConflictResolutionDecision{
			CandidateSkillID: "go-docs-skill",
			TargetSkillID:    "go-docs-skill",
			Mode:             overlap.ResolutionUpdate,
			Blocking:         false,
			SelectedAt:       &selectedAt,
		},
	}
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(input string) string {
	return ansiPattern.ReplaceAllString(input, "")
}
