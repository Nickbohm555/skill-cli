package validation

import "strings"

type SkillMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Extra       map[string]string `json:"extra,omitempty"`
}

type TextSection struct {
	Heading string `json:"heading"`
	Body    string `json:"body,omitempty"`
}

type ListSection struct {
	Heading string   `json:"heading"`
	Intro   string   `json:"intro,omitempty"`
	Items   []string `json:"items,omitempty"`
}

type CandidateSkill struct {
	Metadata        SkillMetadata `json:"metadata"`
	Title           string        `json:"title,omitempty"`
	PurposeSummary  TextSection   `json:"purpose_summary"`
	PrimaryTasks    ListSection   `json:"primary_tasks"`
	SuccessCriteria ListSection   `json:"success_criteria"`
	Constraints     ListSection   `json:"constraints"`
	Dependencies    ListSection   `json:"dependencies"`
	ExampleRequests ListSection   `json:"example_requests"`
	ExampleOutputs  ListSection   `json:"example_outputs"`
	InScope         ListSection   `json:"in_scope"`
	OutOfScope      ListSection   `json:"out_of_scope"`
}

type sectionBinding struct {
	kind sectionKind
	text *TextSection
	list *ListSection
}

type sectionKind int

const (
	sectionKindText sectionKind = iota + 1
	sectionKindList
)

var sectionBindings = map[string]sectionBinding{
	"purpose":          {kind: sectionKindText},
	"purpose-summary":  {kind: sectionKindText},
	"summary":          {kind: sectionKindText},
	"primary-tasks":    {kind: sectionKindList},
	"tasks":            {kind: sectionKindList},
	"success-criteria": {kind: sectionKindList},
	"constraints":      {kind: sectionKindList},
	"dependencies":     {kind: sectionKindList},
	"example-requests": {kind: sectionKindList},
	"example-prompts":  {kind: sectionKindList},
	"example-outputs":  {kind: sectionKindList},
	"in-scope":         {kind: sectionKindList},
	"scope":            {kind: sectionKindList},
	"out-of-scope":     {kind: sectionKindList},
	"non-goals":        {kind: sectionKindList},
}

func newCandidateSkill() CandidateSkill {
	return CandidateSkill{
		PurposeSummary: TextSection{Heading: "Purpose"},
		PrimaryTasks:   ListSection{Heading: "Primary Tasks"},
		SuccessCriteria: ListSection{
			Heading: "Success Criteria",
		},
		Constraints:     ListSection{Heading: "Constraints"},
		Dependencies:    ListSection{Heading: "Dependencies"},
		ExampleRequests: ListSection{Heading: "Example Requests"},
		ExampleOutputs:  ListSection{Heading: "Example Outputs"},
		InScope:         ListSection{Heading: "In Scope"},
		OutOfScope:      ListSection{Heading: "Out Of Scope"},
	}
}

func (c *CandidateSkill) bindSection(slug string) (sectionBinding, bool) {
	binding, ok := sectionBindings[slug]
	if !ok {
		return sectionBinding{}, false
	}

	switch slug {
	case "purpose", "purpose-summary", "summary":
		binding.text = &c.PurposeSummary
	case "primary-tasks", "tasks":
		binding.list = &c.PrimaryTasks
	case "success-criteria":
		binding.list = &c.SuccessCriteria
	case "constraints":
		binding.list = &c.Constraints
	case "dependencies":
		binding.list = &c.Dependencies
	case "example-requests", "example-prompts":
		binding.list = &c.ExampleRequests
	case "example-outputs":
		binding.list = &c.ExampleOutputs
	case "in-scope", "scope":
		binding.list = &c.InScope
	case "out-of-scope", "non-goals":
		binding.list = &c.OutOfScope
	}

	return binding, true
}

func normalizeLines(lines []string) []string {
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}
