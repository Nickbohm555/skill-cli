package overlap

import (
	"math"
	"reflect"
	"testing"
)

func TestDetectNoOverlapReturnsEmptyReport(t *testing.T) {
	t.Parallel()

	candidate := SkillProfile{
		ID:          "candidate.alpha",
		Name:        "alpha docs helper",
		Description: "extracts alpha product documentation",
		InScope:     []string{"collect alpha docs"},
		Commands:    []string{"cli-skill process --url https://alpha.example.com/docs"},
		SourcePath:  "/tmp/candidate/SKILL.md",
	}
	index := InstalledIndex{
		Profiles: []SkillProfile{
			{
				ID:          "installed.beta",
				Name:        "beta runbook helper",
				Description: "summarizes beta incident procedures",
				InScope:     []string{"review beta alerts"},
				Commands:    []string{"cli-skill process --url https://beta.example.com/alerts"},
				SourcePath:  "/Users/test/.codex/skills/beta/SKILL.md",
			},
		},
	}

	report := Detect(candidate, index)

	if report.OverallSeverity != SeverityNone {
		t.Fatalf("OverallSeverity = %q, want %q", report.OverallSeverity, SeverityNone)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("findings = %#v, want none", report.Findings)
	}
}

func TestDetectClassifiesMediumOverlapWithSignals(t *testing.T) {
	t.Parallel()

	candidate := SkillProfile{
		ID:         "candidate.docs",
		Name:       "fresh docs builder",
		InScope:    []string{"extract docs guidance"},
		Commands:   []string{"cli-skill process --url https://example.com/docs"},
		SourcePath: "/tmp/candidate/SKILL.md",
	}
	index := InstalledIndex{
		Profiles: []SkillProfile{
			{
				ID:         "installed.docs",
				Name:       "existing skill helper",
				InScope:    []string{"extract docs guidance"},
				Commands:   []string{"cli-skill process --url https://example.com/docs"},
				SourcePath: "/Users/test/.codex/skills/docs/SKILL.md",
			},
		},
	}

	report := Detect(candidate, index)
	if report.OverallSeverity != SeverityMedium {
		t.Fatalf("OverallSeverity = %q, want %q", report.OverallSeverity, SeverityMedium)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("findings len = %d, want 1", len(report.Findings))
	}

	finding := report.Findings[0]
	if finding.RuleID != "OVLP.STRUCTURAL.OVERLAP" {
		t.Fatalf("RuleID = %q", finding.RuleID)
	}
	if finding.Severity != SeverityMedium {
		t.Fatalf("Severity = %q, want %q", finding.Severity, SeverityMedium)
	}
	if math.Abs(finding.Score-0.45) > 1e-9 {
		t.Fatalf("Score = %.12f, want 0.45", finding.Score)
	}

	gotSignalKeys := []string{finding.Signals[0].Key, finding.Signals[1].Key}
	wantSignalKeys := []string{"in_scope_jaccard", "command_overlap"}
	if !reflect.DeepEqual(gotSignalKeys, wantSignalKeys) {
		t.Fatalf("signal keys = %#v, want %#v", gotSignalKeys, wantSignalKeys)
	}
}

func TestDetectClassifiesHighExactNameCollision(t *testing.T) {
	t.Parallel()

	candidate := SkillProfile{
		ID:         "candidate.docs",
		Name:       "docs helper",
		SourcePath: "/tmp/candidate/SKILL.md",
	}
	index := InstalledIndex{
		Profiles: []SkillProfile{
			{
				ID:         "installed.docs",
				Name:       "Docs Helper",
				SourcePath: "/Users/test/.codex/skills/docs/SKILL.md",
			},
		},
	}

	report := Detect(candidate, index)
	if report.OverallSeverity != SeverityHigh {
		t.Fatalf("OverallSeverity = %q, want %q", report.OverallSeverity, SeverityHigh)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("findings len = %d, want 1", len(report.Findings))
	}
	if report.Findings[0].RuleID != "OVLP.NAME.EXACT" {
		t.Fatalf("RuleID = %q, want OVLP.NAME.EXACT", report.Findings[0].RuleID)
	}
	if report.Findings[0].Signals[0].Key != "name_exact" {
		t.Fatalf("signal key = %q, want name_exact", report.Findings[0].Signals[0].Key)
	}
}

func TestDetectClassifiesExactContentMatch(t *testing.T) {
	t.Parallel()

	candidate := SkillProfile{
		ID:          "candidate.docs",
		Name:        "Docs Helper",
		Description: "Generate a focused skill from docs.",
		InScope:     []string{"Extract docs guidance"},
		OutOfScope:  []string{"Install into codex home"},
		Commands:    []string{"cli-skill process --url https://example.com/docs"},
		SourcePath:  "/tmp/candidate/SKILL.md",
	}
	index := InstalledIndex{
		Profiles: []SkillProfile{
			{
				ID:          "installed.docs",
				Name:        " docs helper ",
				Description: " generate a focused skill from docs. ",
				InScope:     []string{"extract docs guidance"},
				OutOfScope:  []string{"install into codex home"},
				Commands:    []string{"cli-skill process --url https://example.com/docs"},
				SourcePath:  "/Users/test/.codex/skills/docs/SKILL.md",
			},
		},
	}

	report := Detect(candidate, index)
	if report.OverallSeverity != SeverityHigh {
		t.Fatalf("OverallSeverity = %q, want %q", report.OverallSeverity, SeverityHigh)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("findings len = %d, want 1", len(report.Findings))
	}

	finding := report.Findings[0]
	if finding.RuleID != "OVLP.CONTENT.EXACT" {
		t.Fatalf("RuleID = %q, want OVLP.CONTENT.EXACT", finding.RuleID)
	}
	if finding.Score != 1 {
		t.Fatalf("Score = %.2f, want 1.00", finding.Score)
	}
	if finding.Signals[0].Key != "content_exact" {
		t.Fatalf("first signal key = %q, want content_exact", finding.Signals[0].Key)
	}
}

func TestDetectOrdersFindingsDeterministically(t *testing.T) {
	t.Parallel()

	candidate := SkillProfile{
		ID:         "candidate.docs",
		Name:       "Docs Helper",
		InScope:    []string{"extract docs guidance"},
		Commands:   []string{"cli-skill process --url https://example.com/docs"},
		SourcePath: "/tmp/candidate/SKILL.md",
	}
	index := InstalledIndex{
		Profiles: []SkillProfile{
			{
				ID:         "installed.medium",
				Name:       "different helper",
				InScope:    []string{"extract docs guidance"},
				Commands:   []string{"cli-skill process --url https://example.com/docs"},
				SourcePath: "/Users/test/.codex/skills/zeta/SKILL.md",
			},
			{
				ID:         "installed.high",
				Name:       "docs helper",
				SourcePath: "/Users/test/.codex/skills/alpha/SKILL.md",
			},
		},
	}

	report := Detect(candidate, index)
	got := []string{report.Findings[0].RuleID, report.Findings[1].RuleID}
	want := []string{"OVLP.NAME.EXACT", "OVLP.STRUCTURAL.OVERLAP"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("finding order = %#v, want %#v", got, want)
	}
}
