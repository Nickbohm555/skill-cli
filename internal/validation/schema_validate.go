package validation

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	ruleStructuralInternal = "VAL.STRUCT.INTERNAL"
)

//go:embed skill.schema.json
var skillSchemaSource string

var (
	structuralSchemaOnce sync.Once
	structuralSchema     *jsonschema.Schema
	structuralSchemaErr  error
)

type structuralRule struct {
	path      string
	ruleID    string
	label     string
	priority  int
	itemLabel string
}

var structuralRuleByPointer = map[string]structuralRule{
	"/metadata/name": {
		path:     "metadata.name",
		ruleID:   "VAL.STRUCT.METADATA_NAME_REQUIRED",
		label:    "metadata.name",
		priority: 10,
	},
	"/metadata/description": {
		path:     "metadata.description",
		ruleID:   "VAL.STRUCT.METADATA_DESCRIPTION_REQUIRED",
		label:    "metadata.description",
		priority: 20,
	},
	"/title": {
		path:     "title",
		ruleID:   "VAL.STRUCT.TITLE_REQUIRED",
		label:    "title",
		priority: 30,
	},
	"/purpose_summary/body": {
		path:     "sections.purpose_summary.body",
		ruleID:   "VAL.STRUCT.PURPOSE_SUMMARY_REQUIRED",
		label:    "Purpose section",
		priority: 40,
	},
	"/primary_tasks/items": {
		path:      "sections.primary_tasks.items",
		ruleID:    "VAL.STRUCT.PRIMARY_TASKS_REQUIRED",
		label:     "Primary Tasks section",
		priority:  50,
		itemLabel: "Primary Tasks entries",
	},
	"/success_criteria/items": {
		path:      "sections.success_criteria.items",
		ruleID:    "VAL.STRUCT.SUCCESS_CRITERIA_REQUIRED",
		label:     "Success Criteria section",
		priority:  60,
		itemLabel: "Success Criteria entries",
	},
	"/constraints/items": {
		path:      "sections.constraints.items",
		ruleID:    "VAL.STRUCT.CONSTRAINTS_REQUIRED",
		label:     "Constraints section",
		priority:  70,
		itemLabel: "Constraints entries",
	},
	"/dependencies/items": {
		path:      "sections.dependencies.items",
		ruleID:    "VAL.STRUCT.DEPENDENCIES_REQUIRED",
		label:     "Dependencies section",
		priority:  80,
		itemLabel: "Dependencies entries",
	},
	"/example_requests/items": {
		path:      "sections.example_requests.items",
		ruleID:    "VAL.STRUCT.EXAMPLE_REQUESTS_REQUIRED",
		label:     "Example Requests section",
		priority:  90,
		itemLabel: "Example Requests entries",
	},
	"/example_outputs/items": {
		path:      "sections.example_outputs.items",
		ruleID:    "VAL.STRUCT.EXAMPLE_OUTPUTS_REQUIRED",
		label:     "Example Outputs section",
		priority:  100,
		itemLabel: "Example Outputs entries",
	},
	"/in_scope/items": {
		path:      "sections.in_scope.items",
		ruleID:    "VAL.STRUCT.IN_SCOPE_REQUIRED",
		label:     "In Scope section",
		priority:  110,
		itemLabel: "In Scope entries",
	},
	"/out_of_scope/items": {
		path:      "sections.out_of_scope.items",
		ruleID:    "VAL.STRUCT.OUT_OF_SCOPE_REQUIRED",
		label:     "Out Of Scope section",
		priority:  120,
		itemLabel: "Out Of Scope entries",
	},
}

func ValidateStructural(candidate CandidateSkill) ValidationReport {
	report := NewReport()

	schema, err := loadStructuralSchema()
	if err != nil {
		report.AddIssue(ValidationIssue{
			RuleID:   ruleStructuralInternal,
			Severity: SeverityError,
			Message:  fmt.Sprintf("validation schema unavailable: %v", err),
			Priority: 1,
		})
		return report
	}

	payload, err := candidateJSON(candidate)
	if err != nil {
		report.AddIssue(ValidationIssue{
			RuleID:   ruleStructuralInternal,
			Severity: SeverityError,
			Message:  fmt.Sprintf("encode candidate skill: %v", err),
			Priority: 1,
		})
		return report
	}

	if err := schema.Validate(payload); err != nil {
		var validationErr *jsonschema.ValidationError
		if !errors.As(err, &validationErr) {
			report.AddIssue(ValidationIssue{
				RuleID:   ruleStructuralInternal,
				Severity: SeverityError,
				Message:  fmt.Sprintf("validate candidate skill: %v", err),
				Priority: 1,
			})
			return report
		}

		report.AddIssues(mapStructuralIssues(validationErr)...)
	}

	return report
}

func loadStructuralSchema() (*jsonschema.Schema, error) {
	structuralSchemaOnce.Do(func() {
		compiler := jsonschema.NewCompiler()
		schemaDoc, err := jsonschema.UnmarshalJSON(strings.NewReader(skillSchemaSource))
		if err != nil {
			structuralSchemaErr = fmt.Errorf("parse schema resource: %w", err)
			return
		}
		if err := compiler.AddResource("skill.schema.json", schemaDoc); err != nil {
			structuralSchemaErr = fmt.Errorf("add schema resource: %w", err)
			return
		}
		structuralSchema, structuralSchemaErr = compiler.Compile("skill.schema.json")
		if structuralSchemaErr != nil {
			structuralSchemaErr = fmt.Errorf("compile schema: %w", structuralSchemaErr)
		}
	})

	return structuralSchema, structuralSchemaErr
}

func candidateJSON(candidate CandidateSkill) (any, error) {
	raw, err := json.Marshal(candidate)
	if err != nil {
		return nil, err
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func mapStructuralIssues(validationErr *jsonschema.ValidationError) []ValidationIssue {
	output := validationErr.DetailedOutput()
	leaves := flattenStructuralOutput(output)
	issues := make([]ValidationIssue, 0, len(leaves))
	seen := make(map[string]struct{})

	for _, leaf := range leaves {
		issue, ok := mapStructuralIssue(leaf)
		if !ok {
			continue
		}

		key := issue.RuleID + "|" + issue.Path + "|" + issue.Message
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		issues = append(issues, issue)
	}

	if len(issues) == 0 {
		issues = append(issues, ValidationIssue{
			RuleID:   ruleStructuralInternal,
			Severity: SeverityError,
			Message:  validationErr.Error(),
			Priority: 1,
		})
	}

	return issues
}

func flattenStructuralOutput(unit *jsonschema.OutputUnit) []jsonschema.OutputUnit {
	if unit == nil {
		return nil
	}

	if len(unit.Errors) == 0 {
		if unit.Error == nil {
			return nil
		}
		return []jsonschema.OutputUnit{*unit}
	}

	leaves := make([]jsonschema.OutputUnit, 0, len(unit.Errors))
	for i := range unit.Errors {
		childLeaves := flattenStructuralOutput(&unit.Errors[i])
		leaves = append(leaves, childLeaves...)
	}
	return leaves
}

func mapStructuralIssue(unit jsonschema.OutputUnit) (ValidationIssue, bool) {
	rule, path, ok := resolveStructuralRule(unit.InstanceLocation, unit.KeywordLocation, structuralErrorMessage(unit))
	if !ok {
		return ValidationIssue{
			RuleID:   ruleStructuralInternal,
			Severity: SeverityError,
			Path:     jsonPointerToReportPath(unit.InstanceLocation),
			Message:  structuralErrorMessage(unit),
			Priority: 999,
		}, true
	}

	return ValidationIssue{
		RuleID:   rule.ruleID,
		Severity: SeverityError,
		Path:     path,
		Message:  structuralRuleMessage(rule, unit.KeywordLocation, structuralErrorMessage(unit), path),
		Priority: rule.priority,
	}, true
}

func resolveStructuralRule(instanceLocation, keywordLocation, message string) (structuralRule, string, bool) {
	pointer := normalizeInstancePointer(instanceLocation)
	if strings.HasSuffix(keywordLocation, "/required") {
		if missing := extractRequiredProperty(message); missing != "" {
			pointer = strings.TrimSuffix(pointer, "/") + "/" + missing
			if pointer == "/" {
				pointer = "/" + missing
			}
		}
	}

	if rule, ok := structuralRuleByPointer[pointer]; ok {
		return rule, rule.path, true
	}

	parentPointer, itemIndex, hasIndex := trimArrayIndex(pointer)
	if !hasIndex {
		return structuralRule{}, "", false
	}

	rule, ok := structuralRuleByPointer[parentPointer]
	if !ok {
		return structuralRule{}, "", false
	}

	path := rule.path + "[" + strconv.Itoa(itemIndex) + "]"
	return rule, path, true
}

func structuralRuleMessage(rule structuralRule, keywordLocation, fallbackMessage, path string) string {
	switch {
	case strings.HasSuffix(keywordLocation, "/required"):
		return fmt.Sprintf("%s is required.", rule.label)
	case strings.HasSuffix(keywordLocation, "/minLength"):
		if rule.itemLabel != "" && strings.Contains(path, "[") {
			return fmt.Sprintf("%s must not be blank.", rule.itemLabel)
		}
		return fmt.Sprintf("%s must not be blank.", rule.label)
	case strings.HasSuffix(keywordLocation, "/minItems"):
		return fmt.Sprintf("%s must include at least one item.", rule.label)
	case strings.HasSuffix(keywordLocation, "/type"):
		if rule.itemLabel != "" && strings.Contains(path, "[") {
			return fmt.Sprintf("%s must be strings.", rule.itemLabel)
		}
		return fmt.Sprintf("%s has an invalid value type.", rule.label)
	default:
		return fallbackMessage
	}
}

func structuralErrorMessage(unit jsonschema.OutputUnit) string {
	if unit.Error == nil {
		return "schema validation failed"
	}
	return unit.Error.String()
}

func normalizeInstancePointer(pointer string) string {
	if pointer == "" {
		return ""
	}
	if strings.HasPrefix(pointer, "/") {
		return pointer
	}
	return "/" + pointer
}

func trimArrayIndex(pointer string) (string, int, bool) {
	idx := strings.LastIndex(pointer, "/")
	if idx <= 0 || idx >= len(pointer)-1 {
		return "", 0, false
	}

	lastSegment := pointer[idx+1:]
	itemIndex, err := strconv.Atoi(lastSegment)
	if err != nil {
		return "", 0, false
	}

	return pointer[:idx], itemIndex, true
}

func extractRequiredProperty(message string) string {
	const marker = "missing property '"
	start := strings.Index(message, marker)
	if start == -1 {
		return ""
	}
	start += len(marker)
	end := strings.Index(message[start:], "'")
	if end == -1 {
		return ""
	}
	return message[start : start+end]
}

func jsonPointerToReportPath(pointer string) string {
	if pointer == "" {
		return ""
	}

	segments := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	for i := range segments {
		segments[i] = strings.ReplaceAll(segments[i], "~1", "/")
		segments[i] = strings.ReplaceAll(segments[i], "~0", "~")
	}
	return strings.Join(segments, ".")
}
