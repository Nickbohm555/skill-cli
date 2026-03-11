package overlap

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Nickbohm555/skill-cli/internal/validation"
)

type InstalledIndex struct {
	Profiles []SkillProfile `json:"profiles"`
	Warnings []IndexWarning `json:"warnings,omitempty"`
}

func DefaultSkillsRoot() (string, error) {
	codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME"))
	if codexHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve user home: %w", err)
		}
		codexHome = filepath.Join(homeDir, ".codex")
	}

	return filepath.Join(codexHome, "skills"), nil
}

func IndexInstalledSkills(root string) (InstalledIndex, error) {
	if strings.TrimSpace(root) == "" {
		resolvedRoot, err := DefaultSkillsRoot()
		if err != nil {
			return InstalledIndex{}, err
		}
		root = resolvedRoot
	}

	cleanRoot := filepath.Clean(root)
	info, err := os.Stat(cleanRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return InstalledIndex{
				Profiles: []SkillProfile{},
				Warnings: []IndexWarning{},
			}, nil
		}
		return InstalledIndex{}, fmt.Errorf("stat skills root: %w", err)
	}
	if !info.IsDir() {
		return InstalledIndex{}, fmt.Errorf("skills root is not a directory: %s", cleanRoot)
	}

	result := InstalledIndex{
		Profiles: []SkillProfile{},
		Warnings: []IndexWarning{},
	}

	skillPaths := make([]string, 0)
	walkErr := filepath.WalkDir(cleanRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Warnings = append(result.Warnings, IndexWarning{
				Path:    normalizeSourcePath(path),
				Message: walkErr.Error(),
			})
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "SKILL.md") {
			skillPaths = append(skillPaths, path)
		}
		return nil
	})
	if walkErr != nil {
		return InstalledIndex{}, fmt.Errorf("walk skills root: %w", walkErr)
	}

	sort.Slice(skillPaths, func(i, j int) bool {
		return normalizePathKey(skillPaths[i]) < normalizePathKey(skillPaths[j])
	})

	for _, skillPath := range skillPaths {
		input, readErr := os.ReadFile(skillPath)
		if readErr != nil {
			result.Warnings = append(result.Warnings, IndexWarning{
				Path:    normalizeSourcePath(skillPath),
				Message: fmt.Sprintf("read skill file: %v", readErr),
			})
			continue
		}

		candidate, parseErr := validation.ParseSkill(input)
		if parseErr != nil {
			result.Warnings = append(result.Warnings, IndexWarning{
				Path:    normalizeSourcePath(skillPath),
				Message: fmt.Sprintf("parse skill file: %v", parseErr),
			})
			continue
		}

		result.Profiles = append(result.Profiles, profileFromCandidate(cleanRoot, skillPath, candidate))
	}

	sort.SliceStable(result.Profiles, func(i, j int) bool {
		left := result.Profiles[i]
		right := result.Profiles[j]
		if left.SourcePath != right.SourcePath {
			return left.SourcePath < right.SourcePath
		}
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		return left.Name < right.Name
	})
	sort.SliceStable(result.Warnings, func(i, j int) bool {
		left := result.Warnings[i]
		right := result.Warnings[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		return left.Message < right.Message
	})

	return result, nil
}

func ProfileFromCandidate(candidate validation.CandidateSkill, sourcePath string) SkillProfile {
	return profileFromCandidate("", sourcePath, candidate)
}

func profileFromCandidate(root, sourcePath string, candidate validation.CandidateSkill) SkillProfile {
	normalizedPath := normalizeSourcePath(sourcePath)
	name := firstNonEmpty(candidate.Metadata.Name, candidate.Title, filepath.Base(filepath.Dir(sourcePath)))
	description := firstNonEmpty(candidate.Metadata.Description, candidate.PurposeSummary.Body)

	return SkillProfile{
		ID:          profileID(root, sourcePath, name),
		Name:        normalizeComparableText(name),
		Description: normalizeComparableText(description),
		InScope:     normalizeList(candidate.InScope.Items),
		OutOfScope:  normalizeList(candidate.OutOfScope.Items),
		Commands:    extractCommands(candidate),
		SourcePath:  normalizedPath,
	}
}

func profileID(root, sourcePath, fallbackName string) string {
	basePath := sourcePath
	if root != "" {
		if rel, err := filepath.Rel(root, sourcePath); err == nil {
			basePath = rel
		}
	}
	basePath = filepath.Clean(basePath)
	if strings.EqualFold(filepath.Base(basePath), "SKILL.md") {
		basePath = filepath.Dir(basePath)
	}
	normalized := normalizeComparableText(strings.ReplaceAll(filepath.ToSlash(basePath), "/", " "))
	if normalized == "" {
		normalized = normalizeComparableText(fallbackName)
	}
	fields := strings.FieldsFunc(normalized, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return false
		case r >= '0' && r <= '9':
			return false
		}
		return true
	})
	if len(fields) == 0 {
		return "skill"
	}
	return strings.Join(fields, ".")
}

func normalizeList(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		cleaned := normalizeComparableText(value)
		if cleaned == "" {
			continue
		}
		if _, exists := seen[cleaned]; exists {
			continue
		}
		seen[cleaned] = struct{}{}
		normalized = append(normalized, cleaned)
	}
	sort.Strings(normalized)
	return normalized
}

func extractCommands(candidate validation.CandidateSkill) []string {
	seen := make(map[string]struct{})
	commands := make([]string, 0)
	for _, segment := range commandSegments(candidate) {
		for _, part := range splitCommandCandidates(segment) {
			normalized := normalizeComparableText(part)
			if !looksLikeCommand(normalized) {
				continue
			}
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			commands = append(commands, normalized)
		}
	}
	sort.Strings(commands)
	return commands
}

func commandSegments(candidate validation.CandidateSkill) []string {
	segments := []string{
		candidate.PurposeSummary.Body,
		candidate.PrimaryTasks.Intro,
		candidate.SuccessCriteria.Intro,
		candidate.Constraints.Intro,
		candidate.Dependencies.Intro,
		candidate.ExampleRequests.Intro,
		candidate.ExampleOutputs.Intro,
	}
	segments = append(segments, candidate.PrimaryTasks.Items...)
	segments = append(segments, candidate.SuccessCriteria.Items...)
	segments = append(segments, candidate.Constraints.Items...)
	segments = append(segments, candidate.Dependencies.Items...)
	segments = append(segments, candidate.ExampleRequests.Items...)
	segments = append(segments, candidate.ExampleOutputs.Items...)
	return segments
}

func splitCommandCandidates(input string) []string {
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ';'
	})
	candidates := make([]string, 0, len(fields))
	for _, field := range fields {
		trimmed := strings.TrimSpace(field)
		if trimmed == "" {
			continue
		}
		candidates = append(candidates, trimmed)
	}
	return candidates
}

func looksLikeCommand(input string) bool {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return false
	}
	head := fields[0]
	switch head {
	case "go":
		if len(fields) < 2 {
			return false
		}
		switch fields[1] {
		case "build", "env", "fmt", "generate", "install", "list", "mod", "run", "test", "tool", "vet", "work":
			return true
		default:
			return false
		}
	case "cli-skill", "npm", "npx", "yarn", "pnpm", "pip", "pip3", "python", "python3", "uv", "curl", "wget", "git", "docker", "kubectl", "make", "bash", "sh", "node", "./cli-skill", "./bin/cli-skill":
		return true
	}
	return strings.Contains(input, " --") && (strings.Contains(head, "/") || strings.Contains(head, ".") || strings.Contains(head, "-"))
}

func normalizeComparableText(input string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(input))), " ")
}

func normalizeSourcePath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func normalizePathKey(path string) string {
	return strings.ToLower(normalizeSourcePath(path))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
