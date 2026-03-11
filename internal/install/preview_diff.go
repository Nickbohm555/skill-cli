package install

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	diffmatchpatch "github.com/sergi/go-diff/diffmatchpatch"

	"github.com/Nickbohm555/skill-cli/internal/overlap"
	"github.com/Nickbohm555/skill-cli/internal/validation"
)

func RenderPreview(request InstallRequest) string {
	candidate := request.Candidate.Skill

	lines := []string{
		"Install Preview",
		fmt.Sprintf("Mode: %s", previewMode(request.ConflictDecision)),
		fmt.Sprintf("Target path: %s", previewTargetPath(request.Target)),
		fmt.Sprintf("Skill ID: %s", previewSkillID(request)),
	}

	if sourcePath := strings.TrimSpace(request.Candidate.SourcePath); sourcePath != "" {
		lines = append(lines, fmt.Sprintf("Candidate source: %s", sourcePath))
	}
	if existingPath := strings.TrimSpace(request.Target.ExistingPath); existingPath != "" {
		lines = append(lines, fmt.Sprintf("Existing source: %s", existingPath))
	}

	lines = append(lines, "")
	lines = append(lines, renderCandidateSummary(candidate)...)

	return strings.Join(lines, "\n")
}

func RenderDiff(request InstallRequest, existingSkill string) string {
	candidateText := RenderCandidateSkillMarkdown(request.Candidate.Skill)
	if strings.TrimSpace(existingSkill) == "" {
		return strings.Join([]string{
			"Install Diff",
			fmt.Sprintf("Mode: %s", previewMode(request.ConflictDecision)),
			fmt.Sprintf("Target path: %s", previewTargetPath(request.Target)),
			"",
			"No existing installed content. Candidate will be added as new content.",
		}, "\n")
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(existingSkill, candidateText, false)
	dmp.DiffCleanupSemantic(diffs)

	return strings.Join([]string{
		"Install Diff",
		fmt.Sprintf("Mode: %s", previewMode(request.ConflictDecision)),
		fmt.Sprintf("Target path: %s", previewTargetPath(request.Target)),
		"",
		strings.TrimSpace(dmp.DiffPrettyText(diffs)),
	}, "\n")
}

func RenderCandidateSkillMarkdown(candidate validation.CandidateSkill) string {
	lines := make([]string, 0, 64)
	lines = append(lines, "---")
	lines = append(lines, fmt.Sprintf("name: %s", strings.TrimSpace(candidate.Metadata.Name)))
	lines = append(lines, fmt.Sprintf("description: %s", strings.TrimSpace(candidate.Metadata.Description)))

	if len(candidate.Metadata.Extra) > 0 {
		lines = append(lines, "metadata:")
		keys := make([]string, 0, len(candidate.Metadata.Extra))
		for key := range candidate.Metadata.Extra {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			lines = append(lines, fmt.Sprintf("  %s: %s", strings.TrimSpace(key), strings.TrimSpace(candidate.Metadata.Extra[key])))
		}
	}

	lines = append(lines, "---", "")

	if title := strings.TrimSpace(candidate.Title); title != "" {
		lines = append(lines, "# "+title, "")
	}

	appendTextSectionMarkdown(&lines, candidate.PurposeSummary)
	appendListSectionMarkdown(&lines, candidate.PrimaryTasks)
	appendListSectionMarkdown(&lines, candidate.SuccessCriteria)
	appendListSectionMarkdown(&lines, candidate.Constraints)
	appendListSectionMarkdown(&lines, candidate.Dependencies)
	appendListSectionMarkdown(&lines, candidate.ExampleRequests)
	appendListSectionMarkdown(&lines, candidate.ExampleOutputs)
	appendListSectionMarkdown(&lines, candidate.InScope)
	appendListSectionMarkdown(&lines, candidate.OutOfScope)

	return strings.TrimSpace(strings.Join(lines, "\n")) + "\n"
}

func renderCandidateSummary(candidate validation.CandidateSkill) []string {
	lines := []string{
		fmt.Sprintf("Name: %s", strings.TrimSpace(candidate.Metadata.Name)),
		fmt.Sprintf("Title: %s", strings.TrimSpace(candidate.Title)),
		fmt.Sprintf("Description: %s", strings.TrimSpace(candidate.Metadata.Description)),
	}

	if len(candidate.Metadata.Extra) > 0 {
		keys := make([]string, 0, len(candidate.Metadata.Extra))
		for key := range candidate.Metadata.Extra {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		lines = append(lines, "Metadata:")
		for _, key := range keys {
			lines = append(lines, fmt.Sprintf("- %s: %s", strings.TrimSpace(key), strings.TrimSpace(candidate.Metadata.Extra[key])))
		}
	}

	appendTextSectionSummary(&lines, candidate.PurposeSummary)
	appendListSectionSummary(&lines, candidate.PrimaryTasks)
	appendListSectionSummary(&lines, candidate.SuccessCriteria)
	appendListSectionSummary(&lines, candidate.Constraints)
	appendListSectionSummary(&lines, candidate.Dependencies)
	appendListSectionSummary(&lines, candidate.ExampleRequests)
	appendListSectionSummary(&lines, candidate.ExampleOutputs)
	appendListSectionSummary(&lines, candidate.InScope)
	appendListSectionSummary(&lines, candidate.OutOfScope)

	return lines
}

func appendTextSectionSummary(lines *[]string, section validation.TextSection) {
	heading := strings.TrimSpace(section.Heading)
	body := strings.TrimSpace(section.Body)
	if heading == "" || body == "" {
		return
	}

	*lines = append(*lines, "", "## "+heading, body)
}

func appendListSectionSummary(lines *[]string, section validation.ListSection) {
	heading := strings.TrimSpace(section.Heading)
	intro := strings.TrimSpace(section.Intro)
	if heading == "" && intro == "" && len(section.Items) == 0 {
		return
	}

	*lines = append(*lines, "", "## "+heading)
	if intro != "" {
		*lines = append(*lines, intro)
	}
	for _, item := range section.Items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		*lines = append(*lines, "- "+trimmed)
	}
}

func appendTextSectionMarkdown(lines *[]string, section validation.TextSection) {
	heading := strings.TrimSpace(section.Heading)
	body := strings.TrimSpace(section.Body)
	if heading == "" || body == "" {
		return
	}

	*lines = append(*lines, "## "+heading, body, "")
}

func appendListSectionMarkdown(lines *[]string, section validation.ListSection) {
	heading := strings.TrimSpace(section.Heading)
	intro := strings.TrimSpace(section.Intro)
	hasItems := false
	for _, item := range section.Items {
		if strings.TrimSpace(item) != "" {
			hasItems = true
			break
		}
	}
	if heading == "" || (!hasItems && intro == "") {
		return
	}

	*lines = append(*lines, "## "+heading)
	if intro != "" {
		*lines = append(*lines, intro)
	}
	for _, item := range section.Items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		*lines = append(*lines, "- "+trimmed)
	}
	*lines = append(*lines, "")
}

func previewMode(decision *overlap.ConflictResolutionDecision) string {
	if decision == nil || strings.TrimSpace(string(decision.Mode)) == "" {
		return string(overlap.ResolutionNewInstall)
	}
	return string(decision.Mode)
}

func previewTargetPath(target InstallTarget) string {
	if skillDir := strings.TrimSpace(target.SkillDir); skillDir != "" {
		return filepath.ToSlash(skillDir)
	}
	if root := strings.TrimSpace(target.RootDir); root != "" && strings.TrimSpace(target.SkillID) != "" {
		return filepath.ToSlash(filepath.Join(root, target.SkillID))
	}
	if existingPath := strings.TrimSpace(target.ExistingPath); existingPath != "" {
		return filepath.ToSlash(filepath.Dir(existingPath))
	}
	return ""
}

func previewSkillID(request InstallRequest) string {
	if skillID := strings.TrimSpace(request.Target.SkillID); skillID != "" {
		return skillID
	}
	if skillID := strings.TrimSpace(request.Candidate.SkillID); skillID != "" {
		return skillID
	}
	if name := strings.TrimSpace(request.Candidate.Skill.Metadata.Name); name != "" {
		return name
	}
	return ""
}
