package refinement

import (
	"fmt"
	"regexp"
	"strings"
)

var reviseCommandPattern = regexp.MustCompile(`^\s*revise\s+([a-z_]+)\s*$`)

type RevisionCommand struct {
	FieldID FieldID
}

func ParseReviseCommand(input string) (RevisionCommand, error) {
	matches := reviseCommandPattern.FindStringSubmatch(input)
	if len(matches) != 2 {
		return RevisionCommand{}, fmt.Errorf("revision command must use strict form: revise <field>")
	}

	return RevisionCommand{FieldID: FieldID(matches[1])}, nil
}

func ValidateRevisionTarget(state *SessionState, fieldID FieldID) (FieldState, error) {
	if state == nil {
		return FieldState{}, fmt.Errorf("revision state is required")
	}

	field, ok := state.Field(fieldID)
	if !ok {
		return FieldState{}, fmt.Errorf("unknown revision field %q", fieldID)
	}
	if !field.Definition.Required {
		return FieldState{}, fmt.Errorf("field %q is not revisable", fieldID)
	}
	if strings.TrimSpace(string(field.Definition.ID)) == "" {
		return FieldState{}, fmt.Errorf("field %q is not revisable", fieldID)
	}

	return field, nil
}
