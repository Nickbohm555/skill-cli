package overlap

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestIndexInstalledSkillsBuildsNormalizedProfiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skillPath := filepath.Join(root, "Tools", "Go Docs", "SKILL.md")
	writeSkillFixture(t, skillPath, `---
name: "  Go DOCS   Helper  "
description: "  Compare one docs URL and generate a focused skill. "
---

# Go Docs Helper

## Purpose

Generate one skill from Go documentation.

## Dependencies

- Go 1.25.x
- go run ./cmd/cli-skill --help

## Example Requests

- cli-skill process --url https://go.dev/doc/
- CLI-SKILL PROCESS --url https://go.dev/doc/

## Example Outputs

- ./bin/cli-skill --help

## In Scope

-  Extract   Go package documentation.
- render SKILL.md   install guidance
- Extract Go Package Documentation.

## Out of Scope

-   Mixing unrelated Docs sources.
- Editing installed skills in place.
`)

	index, err := IndexInstalledSkills(root)
	if err != nil {
		t.Fatalf("IndexInstalledSkills() error = %v", err)
	}

	if len(index.Warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", index.Warnings)
	}
	if len(index.Profiles) != 1 {
		t.Fatalf("profiles len = %d, want 1", len(index.Profiles))
	}

	got := index.Profiles[0]
	if got.ID != "tools.go.docs" {
		t.Fatalf("profile ID = %q, want %q", got.ID, "tools.go.docs")
	}
	if got.Name != "go docs helper" {
		t.Fatalf("profile name = %q, want %q", got.Name, "go docs helper")
	}
	if got.Description != "compare one docs url and generate a focused skill." {
		t.Fatalf("profile description = %q", got.Description)
	}
	if got.SourcePath != filepath.ToSlash(skillPath) {
		t.Fatalf("profile source path = %q, want %q", got.SourcePath, filepath.ToSlash(skillPath))
	}

	wantInScope := []string{
		"extract go package documentation.",
		"render skill.md install guidance",
	}
	if !reflect.DeepEqual(got.InScope, wantInScope) {
		t.Fatalf("in_scope = %#v, want %#v", got.InScope, wantInScope)
	}

	wantOutOfScope := []string{
		"editing installed skills in place.",
		"mixing unrelated docs sources.",
	}
	if !reflect.DeepEqual(got.OutOfScope, wantOutOfScope) {
		t.Fatalf("out_of_scope = %#v, want %#v", got.OutOfScope, wantOutOfScope)
	}

	wantCommands := []string{
		"./bin/cli-skill --help",
		"cli-skill process --url https://go.dev/doc/",
		"go run ./cmd/cli-skill --help",
	}
	if !reflect.DeepEqual(got.Commands, wantCommands) {
		t.Fatalf("commands = %#v, want %#v", got.Commands, wantCommands)
	}
}

func TestIndexInstalledSkillsWarnsOnMalformedSkillFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeSkillFixture(t, filepath.Join(root, "valid", "SKILL.md"), `---
name: valid-skill
description: valid description
---

# Valid Skill

## In Scope

- Extract docs.
`)
	writeSkillFixture(t, filepath.Join(root, "broken", "SKILL.md"), `---
name: [broken
description: missing bracket
---

# Broken Skill
`)

	index, err := IndexInstalledSkills(root)
	if err != nil {
		t.Fatalf("IndexInstalledSkills() error = %v", err)
	}

	if len(index.Profiles) != 1 {
		t.Fatalf("profiles len = %d, want 1", len(index.Profiles))
	}
	if len(index.Warnings) != 1 {
		t.Fatalf("warnings len = %d, want 1", len(index.Warnings))
	}
	if index.Warnings[0].Path != filepath.ToSlash(filepath.Join(root, "broken", "SKILL.md")) {
		t.Fatalf("warning path = %q", index.Warnings[0].Path)
	}
	if !strings.Contains(index.Warnings[0].Message, "parse skill file:") {
		t.Fatalf("warning message = %q, want parse prefix", index.Warnings[0].Message)
	}
}

func TestIndexInstalledSkillsUsesDefaultCodexHomeRoot(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)
	skillPath := filepath.Join(codexHome, "skills", "CaseTest", "skill.md")
	writeSkillFixture(t, skillPath, `---
name: Case Test
description: Case normalization coverage
---

# Case Test

## In Scope

- Compare installed skills carefully.
`)

	index, err := IndexInstalledSkills("")
	if err != nil {
		t.Fatalf("IndexInstalledSkills(\"\") error = %v", err)
	}

	if len(index.Warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", index.Warnings)
	}
	if len(index.Profiles) != 1 {
		t.Fatalf("profiles len = %d, want 1", len(index.Profiles))
	}

	got := index.Profiles[0]
	if got.ID != "casetest" {
		t.Fatalf("profile ID = %q, want %q", got.ID, "casetest")
	}
	if got.SourcePath != filepath.ToSlash(skillPath) {
		t.Fatalf("profile source path = %q, want %q", got.SourcePath, filepath.ToSlash(skillPath))
	}
	if got.Name != "case test" {
		t.Fatalf("profile name = %q, want %q", got.Name, "case test")
	}
}

func writeSkillFixture(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
