package validation

import "strings"

const unknownRulePrompt = "Revise the content flagged by validation so it explicitly satisfies the reported requirement before continuing."

var followupPromptByRule = map[string]string{
	"VAL.STRUCT.INTERNAL": "Fix the generated skill content so the validation failure is resolved, then retry validation.",

	"VAL.STRUCT.METADATA_NAME_REQUIRED":        "Provide a short, specific skill name in the frontmatter `name` field.",
	"VAL.STRUCT.METADATA_DESCRIPTION_REQUIRED": "Provide a concise frontmatter `description` that states what the skill does and what source it uses.",
	"VAL.STRUCT.TITLE_REQUIRED":                "Add the main `#` title for the skill so the document has a clear heading.",
	"VAL.STRUCT.PURPOSE_SUMMARY_REQUIRED":      "Write the `Purpose` section with a brief summary of the skill's goal from the docs URL.",
	"VAL.STRUCT.PRIMARY_TASKS_REQUIRED":        "List the concrete steps or jobs this skill must perform in the `Primary Tasks` section.",
	"VAL.STRUCT.SUCCESS_CRITERIA_REQUIRED":     "List the measurable checks that define when the skill output is acceptable in `Success Criteria`.",
	"VAL.STRUCT.CONSTRAINTS_REQUIRED":          "List the hard limits or rules the skill must obey in `Constraints`.",
	"VAL.STRUCT.DEPENDENCIES_REQUIRED":         "List the required tools, runtimes, or integrations in `Dependencies`.",
	"VAL.STRUCT.EXAMPLE_REQUESTS_REQUIRED":     "Add example user requests that this skill should handle well in `Example Requests`.",
	"VAL.STRUCT.EXAMPLE_OUTPUTS_REQUIRED":      "Add concrete example outputs or deliverables in `Example Outputs`.",
	"VAL.STRUCT.IN_SCOPE_REQUIRED":             "List the specific capabilities or source boundaries that are allowed in `In Scope`.",
	"VAL.STRUCT.OUT_OF_SCOPE_REQUIRED":         "List the explicit exclusions or non-goals in `Out Of Scope`.",

	"VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF":     "Rewrite the flagged `In Scope` item so it names a concrete supported capability or boundary in specific terms.",
	"VAL.SCOPE.IN_SCOPE_VAGUE_CATCH_ALL":     "Rewrite the flagged `In Scope` item to remove vague catch-all wording and state the exact supported boundary.",
	"VAL.SCOPE.OUT_OF_SCOPE_ENTRY_TOO_BRIEF": "Rewrite the flagged `Out Of Scope` item so it names a concrete excluded capability or boundary in specific terms.",
	"VAL.SCOPE.OUT_OF_SCOPE_VAGUE_CATCH_ALL": "Rewrite the flagged `Out Of Scope` item to remove vague catch-all wording and state the exact exclusion.",
}

func PromptForRule(ruleID string) string {
	prompt, ok := followupPromptByRule[strings.TrimSpace(ruleID)]
	if !ok {
		return unknownRulePrompt
	}
	return prompt
}
