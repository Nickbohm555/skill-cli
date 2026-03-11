package prompts

import (
	"fmt"
	"strings"

	huh "charm.land/huh/v2"

	"github.com/Nickbohm555/skill-cli/internal/refinement"
)

const OtherOptionValue = "other"

type PromptKind string

const (
	PromptKindPrimary           PromptKind = "primary"
	PromptKindDeepeningFreeText PromptKind = "deepening_free_text"
	PromptKindDeepeningSelect   PromptKind = "deepening_select"
	PromptKindDeepeningFallback PromptKind = "deepening_fallback"
	PromptKindNoop              PromptKind = "noop"
	defaultSelectHeight                    = 4
)

type ControlType string

const (
	ControlTypeInput  ControlType = "input"
	ControlTypeSelect ControlType = "select"
)

type PromptOption struct {
	Label string
	Value string
}

type QuestionSpec struct {
	Key         string
	FieldID     refinement.FieldID
	Kind        PromptKind
	Control     ControlType
	Title       string
	Description string
	Placeholder string
	Options     []PromptOption
	Required    bool
}

type PromptPlan struct {
	FieldID      refinement.FieldID
	Section      refinement.SectionID
	Label        string
	Kind         PromptKind
	Decision     refinement.DeepeningDecision
	RequireOther bool
	Prompts      []QuestionSpec
}

type DeepeningBindings struct {
	Answer *string
	Choice *string
	Other  *string
}

type RefinementFormAdapter struct {
	policy refinement.ClarityPolicy
}

type fieldPromptContent struct {
	primaryPrompt        string
	primaryPlaceholder   string
	deepeningPrompt      string
	deepeningPlaceholder string
	options              []PromptOption
}

func DefaultRefinementFormAdapter() RefinementFormAdapter {
	return RefinementFormAdapter{policy: refinement.DefaultClarityPolicy()}
}

func NewRefinementFormAdapter(policy refinement.ClarityPolicy) RefinementFormAdapter {
	return RefinementFormAdapter{policy: policy}
}

func (a RefinementFormAdapter) NeedsDeepening(field refinement.FieldState, attempts int) (bool, refinement.DeepeningDecision, error) {
	decision, err := a.policy.DeepeningDecision(field.Definition.ID, field.Answer.Value, attempts)
	if err != nil {
		return false, refinement.DeepeningDecision{}, err
	}
	return decision.Mode != refinement.DeepeningModeNone, decision, nil
}

func MaxAttempts(decision refinement.DeepeningDecision) bool {
	return decision.Mode == refinement.DeepeningModeCapped || decision.Attempt >= decision.MaxAttempts
}

func (a RefinementFormAdapter) PrimaryPlan(field refinement.FieldState) (PromptPlan, error) {
	content, err := promptContent(field.Definition.ID)
	if err != nil {
		return PromptPlan{}, err
	}

	spec := QuestionSpec{
		Key:         questionKey(field.Definition.ID),
		FieldID:     field.Definition.ID,
		Kind:        PromptKindPrimary,
		Control:     ControlTypeInput,
		Title:       field.Definition.Label,
		Description: primaryDescription(content.primaryPrompt),
		Placeholder: content.primaryPlaceholder,
		Required:    field.Definition.Required,
	}

	return PromptPlan{
		FieldID: field.Definition.ID,
		Section: field.Definition.Section,
		Label:   field.Definition.Label,
		Kind:    PromptKindPrimary,
		Prompts: []QuestionSpec{spec},
	}, nil
}

func (a RefinementFormAdapter) DeepeningPlan(field refinement.FieldState, attempts int) (PromptPlan, error) {
	content, err := promptContent(field.Definition.ID)
	if err != nil {
		return PromptPlan{}, err
	}

	decision, err := a.policy.DeepeningDecision(field.Definition.ID, field.Answer.Value, attempts)
	if err != nil {
		return PromptPlan{}, err
	}

	plan := PromptPlan{
		FieldID:      field.Definition.ID,
		Section:      field.Definition.Section,
		Label:        field.Definition.Label,
		Decision:     decision,
		RequireOther: decision.RequireExplicitOther,
	}

	switch decision.Mode {
	case refinement.DeepeningModeNone:
		plan.Kind = PromptKindNoop
		return plan, nil
	case refinement.DeepeningModeFreeText:
		plan.Kind = PromptKindDeepeningFreeText
		plan.Prompts = []QuestionSpec{{
			Key:         questionKey(field.Definition.ID),
			FieldID:     field.Definition.ID,
			Kind:        PromptKindDeepeningFreeText,
			Control:     ControlTypeInput,
			Title:       field.Definition.Label,
			Description: deepeningDescription(content.deepeningPrompt, false),
			Placeholder: content.deepeningPlaceholder,
			Required:    field.Definition.Required,
		}}
		return plan, nil
	case refinement.DeepeningModeStructuredChoice, refinement.DeepeningModeCapped:
		plan.Kind = PromptKindDeepeningSelect
		if decision.Mode == refinement.DeepeningModeCapped {
			plan.Kind = PromptKindDeepeningFallback
		}

		options := append([]PromptOption(nil), content.options...)
		options = append(options, PromptOption{Label: "Other (describe)", Value: OtherOptionValue})

		plan.Prompts = []QuestionSpec{
			{
				Key:         choiceKey(field.Definition.ID),
				FieldID:     field.Definition.ID,
				Kind:        plan.Kind,
				Control:     ControlTypeSelect,
				Title:       field.Definition.Label,
				Description: deepeningDescription(content.deepeningPrompt, MaxAttempts(decision)),
				Options:     options,
				Required:    field.Definition.Required,
			},
			{
				Key:         otherKey(field.Definition.ID),
				FieldID:     field.Definition.ID,
				Kind:        plan.Kind,
				Control:     ControlTypeInput,
				Title:       fmt.Sprintf("%s (other details)", field.Definition.Label),
				Description: otherDescription(field.Definition.Label),
				Placeholder: content.deepeningPlaceholder,
				Required:    false,
			},
		}
		return plan, nil
	default:
		return PromptPlan{}, fmt.Errorf("unsupported deepening mode %q", decision.Mode)
	}
}

func (a RefinementFormAdapter) BuildPrimaryFields(field refinement.FieldState, answer *string) (PromptPlan, []huh.Field, error) {
	plan, err := a.PrimaryPlan(field)
	if err != nil {
		return PromptPlan{}, nil, err
	}
	fields, err := buildFields(plan, DeepeningBindings{Answer: answer})
	if err != nil {
		return PromptPlan{}, nil, err
	}
	return plan, fields, nil
}

func (a RefinementFormAdapter) BuildDeepeningFields(field refinement.FieldState, attempts int, bindings DeepeningBindings) (PromptPlan, []huh.Field, error) {
	plan, err := a.DeepeningPlan(field, attempts)
	if err != nil {
		return PromptPlan{}, nil, err
	}
	fields, err := buildFields(plan, bindings)
	if err != nil {
		return PromptPlan{}, nil, err
	}
	return plan, fields, nil
}

func buildFields(plan PromptPlan, bindings DeepeningBindings) ([]huh.Field, error) {
	fields := make([]huh.Field, 0, len(plan.Prompts))
	for _, spec := range plan.Prompts {
		switch spec.Control {
		case ControlTypeInput:
			target := bindings.Answer
			if spec.Key == otherKey(plan.FieldID) {
				target = bindings.Other
			}
			if target == nil {
				return nil, fmt.Errorf("missing input binding for %q", spec.Key)
			}
			input := huh.NewInput().
				Key(spec.Key).
				Title(spec.Title).
				Description(spec.Description).
				Placeholder(spec.Placeholder).
				Value(target)
			if spec.Required {
				input.Validate(requiredText(spec.Title))
			} else if spec.Key == otherKey(plan.FieldID) {
				input.Validate(requiredOther(plan.FieldID, bindings.Choice))
			}
			fields = append(fields, input)
		case ControlTypeSelect:
			if bindings.Choice == nil {
				return nil, fmt.Errorf("missing select binding for %q", spec.Key)
			}
			options := make([]huh.Option[string], 0, len(spec.Options))
			for _, option := range spec.Options {
				options = append(options, huh.NewOption(option.Label, option.Value))
			}
			selectField := huh.NewSelect[string]().
				Key(spec.Key).
				Title(spec.Title).
				Description(spec.Description).
				Options(options...).
				Height(defaultSelectHeight).
				Value(bindings.Choice)
			if spec.Required {
				selectField.Validate(requiredChoice(spec.Title))
			}
			fields = append(fields, selectField)
		default:
			return nil, fmt.Errorf("unsupported control type %q", spec.Control)
		}
	}
	return fields, nil
}

func requiredText(label string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", label)
		}
		return nil
	}
}

func requiredChoice(label string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("choose one option for %s", label)
		}
		return nil
	}
}

func requiredOther(fieldID refinement.FieldID, selected *string) func(string) error {
	return func(value string) error {
		if selected == nil || strings.TrimSpace(*selected) != OtherOptionValue {
			return nil
		}
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s requires additional detail", fieldID)
		}
		return nil
	}
}

func primaryDescription(prompt string) string {
	return prompt + " Use concrete details so the review can mark this field ready."
}

func deepeningDescription(prompt string, finalAttempt bool) string {
	if finalAttempt {
		return prompt + " Pick the closest structured option or use Other to give the exact detail needed before review."
	}
	return prompt + " Add the missing specificity so this field can move from needs attention to ready."
}

func otherDescription(label string) string {
	return fmt.Sprintf("Use this only when \"%s\" needs a custom answer beyond the listed options.", label)
}

func questionKey(fieldID refinement.FieldID) string {
	return fmt.Sprintf("%s_answer", fieldID)
}

func choiceKey(fieldID refinement.FieldID) string {
	return fmt.Sprintf("%s_choice", fieldID)
}

func otherKey(fieldID refinement.FieldID) string {
	return fmt.Sprintf("%s_other", fieldID)
}

func promptContent(fieldID refinement.FieldID) (fieldPromptContent, error) {
	content, ok := promptCatalog[fieldID]
	if !ok {
		return fieldPromptContent{}, fmt.Errorf("missing prompt content for field %q", fieldID)
	}
	return content, nil
}

var promptCatalog = map[refinement.FieldID]fieldPromptContent{
	refinement.FieldPurposeSummary: {
		primaryPrompt:        "Describe what the skill should accomplish from one documentation URL.",
		primaryPlaceholder:   "Turn one docs site into a Codex skill with install steps, scope limits, and example requests.",
		deepeningPrompt:      "Clarify the exact outcome or focus that is still vague in the purpose summary.",
		deepeningPlaceholder: "Add the concrete outcome, audience, or scope signal that is missing.",
		options: []PromptOption{
			{Label: "Core outcome and user goal", Value: "core_outcome"},
			{Label: "Target audience or usage context", Value: "audience"},
			{Label: "Scope boundary or exclusion", Value: "scope_boundary"},
		},
	},
	refinement.FieldPrimaryTasks: {
		primaryPrompt:        "List the main jobs the generated skill must perform.",
		primaryPlaceholder:   "Fetch the docs, extract relevant guidance, and produce a focused SKILL.md.",
		deepeningPrompt:      "Clarify which concrete actions or workflow steps are missing from the task list.",
		deepeningPlaceholder: "Name the missing action, output, or sequencing detail.",
		options: []PromptOption{
			{Label: "Acquisition and extraction steps", Value: "acquisition_steps"},
			{Label: "Transformation into skill instructions", Value: "transformation_steps"},
			{Label: "Review or validation steps", Value: "review_steps"},
		},
	},
	refinement.FieldSuccessCriteria: {
		primaryPrompt:        "State how you will know the generated skill is good enough to ship.",
		primaryPlaceholder:   "The skill installs cleanly, stays scoped to one docs set, and includes realistic examples.",
		deepeningPrompt:      "Clarify the measurable quality bar or acceptance criteria that are still underspecified.",
		deepeningPlaceholder: "Add a concrete pass condition, constraint, or quality check.",
		options: []PromptOption{
			{Label: "Correct install and run behavior", Value: "install_quality"},
			{Label: "Scope and safety boundaries", Value: "scope_quality"},
			{Label: "Example or output quality", Value: "example_quality"},
		},
	},
	refinement.FieldConstraints: {
		primaryPrompt:        "List the hard constraints the generator must obey.",
		primaryPlaceholder:   "Use one docs URL only, avoid speculation, and keep the output deterministic.",
		deepeningPrompt:      "Clarify which hard limits or operating constraints need to be stated more explicitly.",
		deepeningPlaceholder: "Add the missing rule, forbidden behavior, or environment constraint.",
		options: []PromptOption{
			{Label: "Input or source limitations", Value: "input_limitations"},
			{Label: "Behavior the skill must avoid", Value: "forbidden_behavior"},
			{Label: "Environment or tooling requirement", Value: "environment_requirement"},
		},
	},
	refinement.FieldDependencies: {
		primaryPrompt:        "Describe the required dependencies, integrations, or runtime assumptions.",
		primaryPlaceholder:   "Needs network access to fetch docs and optional OpenAI credentials for structured summaries.",
		deepeningPrompt:      "Clarify which dependency or runtime assumption is missing from the answer.",
		deepeningPlaceholder: "Add the missing tool, credential, service, or environment assumption.",
		options: []PromptOption{
			{Label: "Network or external service dependency", Value: "network_dependency"},
			{Label: "Credential or secret requirement", Value: "credential_requirement"},
			{Label: "Local toolchain or environment need", Value: "toolchain_requirement"},
		},
	},
	refinement.FieldExampleRequests: {
		primaryPrompt:        "Provide example user requests the skill should handle well.",
		primaryPlaceholder:   "Generate a skill from Go docs and tighten the scope when tutorials and API refs are mixed.",
		deepeningPrompt:      "Clarify the missing request pattern or scenario detail in the examples.",
		deepeningPlaceholder: "Add a realistic request, trigger condition, or edge case.",
		options: []PromptOption{
			{Label: "Happy-path request", Value: "happy_path"},
			{Label: "Revision or refinement request", Value: "revision_request"},
			{Label: "Failure or edge-case request", Value: "edge_case"},
		},
	},
	refinement.FieldExampleOutputs: {
		primaryPrompt:        "Describe what the generated skill output should contain in concrete terms.",
		primaryPlaceholder:   "Show install commands, supported inputs, and one explicit out-of-scope case.",
		deepeningPrompt:      "Clarify which expected output details or formatting cues are still missing.",
		deepeningPlaceholder: "Add the missing output element, format, or example detail.",
		options: []PromptOption{
			{Label: "Install or setup details", Value: "install_output"},
			{Label: "Usage examples or commands", Value: "usage_output"},
			{Label: "Warnings, exclusions, or guardrails", Value: "guardrail_output"},
		},
	},
	refinement.FieldInScope: {
		primaryPrompt:        "State the work that is explicitly in scope for the skill.",
		primaryPlaceholder:   "In scope: extract instructions, commands, and examples from the chosen docs set.",
		deepeningPrompt:      "Clarify which supported capabilities or source boundaries should be named more explicitly.",
		deepeningPlaceholder: "Add the supported capability, content type, or source boundary.",
		options: []PromptOption{
			{Label: "Supported content or document types", Value: "supported_content"},
			{Label: "Allowed operations or outputs", Value: "allowed_operations"},
			{Label: "Source boundary or coverage limit", Value: "source_boundary"},
		},
	},
	refinement.FieldOutOfScope: {
		primaryPrompt:        "State the work that must stay out of scope for the skill.",
		primaryPlaceholder:   "Out of scope: inventing APIs, mixing unrelated docs sets, or building extra tooling.",
		deepeningPrompt:      "Clarify which exclusions or non-goals need to be stated more concretely.",
		deepeningPlaceholder: "Add the excluded behavior, source, or output that should be blocked.",
		options: []PromptOption{
			{Label: "Disallowed source expansion", Value: "source_exclusion"},
			{Label: "Unsupported workflow or capability", Value: "workflow_exclusion"},
			{Label: "Speculative or invented content", Value: "speculation_exclusion"},
		},
	},
}
