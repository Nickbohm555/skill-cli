package command

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Nickbohm555/skill-cli/internal/cli/prompts"
	"github.com/Nickbohm555/skill-cli/internal/refinement"
	"github.com/spf13/cobra"
)

type refinementPayload struct {
	Phase       string                    `json:"phase"`
	State       refinement.FlowState      `json:"state"`
	CommitReady bool                      `json:"commit_ready"`
	Revision    int                       `json:"revision"`
	Answers     []refinementPayloadAnswer `json:"answers"`
}

type refinementPayloadAnswer struct {
	FieldID  refinement.FieldID         `json:"field_id"`
	Section  refinement.SectionID       `json:"section"`
	Label    string                     `json:"label"`
	Value    string                     `json:"value"`
	Status   refinement.ReadinessStatus `json:"status"`
	Revision int                        `json:"revision"`
}

type refinementConsole struct {
	reader  *bufio.Reader
	out     io.Writer
	errOut  io.Writer
	adapter prompts.RefinementFormAdapter
}

func newRefineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refine",
		Short: "Collect and refine required skill inputs before generation",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := runRefineSession(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
			if err != nil {
				return fmt.Errorf("refine failed: %w", err)
			}
			return nil
		},
	}

	return cmd
}

func runRefineSession(in io.Reader, out io.Writer, errOut io.Writer) (refinementPayload, error) {
	state, err := refinement.NewSessionState()
	if err != nil {
		return refinementPayload{}, err
	}

	console := refinementConsole{
		reader:  bufio.NewReader(in),
		out:     out,
		errOut:  errOut,
		adapter: prompts.DefaultRefinementFormAdapter(),
	}

	flow, err := refinement.NewFlow(state, &console, &console)
	if err != nil {
		return refinementPayload{}, err
	}

	result, err := flow.Run()
	if err != nil {
		return refinementPayload{}, err
	}
	if result.State != refinement.FlowStateReview {
		return refinementPayload{}, fmt.Errorf("refinement did not reach review state")
	}

	for {
		console.renderReview(result.Report)

		command, err := console.readLine("review> ")
		if err != nil {
			return refinementPayload{}, err
		}
		command = strings.TrimSpace(command)
		if command == "" || strings.EqualFold(command, "review") {
			continue
		}

		switch {
		case strings.EqualFold(command, "commit"):
			committed, commitErr := flow.Commit()
			if commitErr != nil {
				if _, writeErr := fmt.Fprintf(console.errOut, "%v\n", commitErr); writeErr != nil {
					return refinementPayload{}, writeErr
				}
				result.Report = flowReportOrFallback(state, result.Report)
				continue
			}

			payload := buildRefinementPayload(state, committed)
			if err := writePayload(out, payload); err != nil {
				return refinementPayload{}, err
			}
			return payload, nil
		case strings.EqualFold(command, "quit"), strings.EqualFold(command, "exit"), strings.EqualFold(command, "abort"):
			return refinementPayload{}, fmt.Errorf("refinement aborted before commit")
		case strings.HasPrefix(command, "revise "):
			revision, parseErr := refinement.ParseReviseCommand(command)
			if parseErr != nil {
				if _, writeErr := fmt.Fprintf(console.errOut, "%v\n", parseErr); writeErr != nil {
					return refinementPayload{}, writeErr
				}
				continue
			}

			field, validateErr := refinement.ValidateRevisionTarget(state, revision.FieldID)
			if validateErr != nil {
				if _, writeErr := fmt.Fprintf(console.errOut, "%v\n", validateErr); writeErr != nil {
					return refinementPayload{}, writeErr
				}
				continue
			}

			answer, askErr := console.askRevision(field)
			if askErr != nil {
				return refinementPayload{}, askErr
			}

			result, err = flow.Revise(command, answer)
			if err != nil {
				return refinementPayload{}, err
			}
		default:
			if _, writeErr := fmt.Fprintln(console.errOut, "unknown review command; use review, revise <field>, commit, or quit"); writeErr != nil {
				return refinementPayload{}, writeErr
			}
		}
	}
}

func (c *refinementConsole) AskPrimary(field refinement.FieldState) (string, error) {
	plan, err := c.adapter.PrimaryPlan(field)
	if err != nil {
		return "", err
	}
	return c.askPlan(plan, field.Answer.Value, false)
}

func (c *refinementConsole) AskDeepening(field refinement.FieldState, decision refinement.DeepeningDecision, signal refinement.SummarizeFirstSignal) (string, error) {
	plan, err := c.adapter.DeepeningPlan(field, decision.Attempt)
	if err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(c.out, "\nSummary before follow-up for %s: %s\n", field.Definition.ID, strings.TrimSpace(signal.Summary)); err != nil {
		return "", err
	}

	return c.askPlan(plan, "", false)
}

func (c *refinementConsole) SummarizeFirst(field refinement.FieldState, decision refinement.DeepeningDecision) (refinement.SummarizeFirstSignal, error) {
	summary := fmt.Sprintf("%s needs more specificity before review; follow-up mode: %s.", field.Definition.Label, decision.Mode)
	return refinement.SummarizeFirstSignal{
		FieldID:  field.Definition.ID,
		Section:  field.Definition.Section,
		Answer:   field.Answer.Value,
		Summary:  summary,
		Decision: decision,
	}, nil
}

func (c *refinementConsole) askRevision(field refinement.FieldState) (string, error) {
	if _, err := fmt.Fprintf(c.out, "\nRevising %s (%s)\n", field.Definition.Label, field.Definition.ID); err != nil {
		return "", err
	}
	if current := strings.TrimSpace(field.Answer.Value); current != "" {
		if _, err := fmt.Fprintf(c.out, "Current answer: %s\n", current); err != nil {
			return "", err
		}
	}

	plan, err := c.adapter.PrimaryPlan(field)
	if err != nil {
		return "", err
	}
	return c.askPlan(plan, field.Answer.Value, true)
}

func (c *refinementConsole) askPlan(plan prompts.PromptPlan, existing string, revision bool) (string, error) {
	if _, err := fmt.Fprintf(c.out, "\n[%s] %s\n", plan.Section, plan.Label); err != nil {
		return "", err
	}

	for _, prompt := range plan.Prompts {
		if prompt.Key == "" {
			continue
		}
		if prompt.Control == prompts.ControlTypeInput && strings.Contains(prompt.Key, "other") {
			continue
		}
		switch prompt.Control {
		case prompts.ControlTypeInput:
			return c.askText(prompt, existing, revision)
		case prompts.ControlTypeSelect:
			return c.askSelect(plan, prompt)
		}
	}

	return "", fmt.Errorf("prompt plan for %q produced no usable prompts", plan.FieldID)
}

func (c *refinementConsole) askText(prompt prompts.QuestionSpec, existing string, revision bool) (string, error) {
	if _, err := fmt.Fprintf(c.out, "%s\n", prompt.Description); err != nil {
		return "", err
	}
	if placeholder := strings.TrimSpace(prompt.Placeholder); placeholder != "" {
		if _, err := fmt.Fprintf(c.out, "Hint: %s\n", placeholder); err != nil {
			return "", err
		}
	}
	if revision && strings.TrimSpace(existing) != "" {
		if _, err := fmt.Fprintf(c.out, "Press enter after replacing the current answer.\n"); err != nil {
			return "", err
		}
	}

	for {
		answer, err := c.readLine("> ")
		if err != nil {
			return "", err
		}
		answer = strings.TrimSpace(answer)
		if prompt.Required && answer == "" {
			if _, err := fmt.Fprintf(c.errOut, "%s is required\n", prompt.FieldID); err != nil {
				return "", err
			}
			continue
		}
		return answer, nil
	}
}

func (c *refinementConsole) askSelect(plan prompts.PromptPlan, prompt prompts.QuestionSpec) (string, error) {
	if _, err := fmt.Fprintf(c.out, "%s\n", prompt.Description); err != nil {
		return "", err
	}
	for index, option := range prompt.Options {
		if _, err := fmt.Fprintf(c.out, "%d. %s\n", index+1, option.Label); err != nil {
			return "", err
		}
	}

	for {
		raw, err := c.readLine("> ")
		if err != nil {
			return "", err
		}

		selected, selectedLabel, ok := resolveOption(prompt.Options, raw)
		if !ok {
			if _, err := fmt.Fprintln(c.errOut, "choose an option by number, label, or value"); err != nil {
				return "", err
			}
			continue
		}

		if selected == prompts.OtherOptionValue {
			otherPrompt, ok := findOtherPrompt(plan)
			if !ok {
				return "", fmt.Errorf("other prompt missing for %q", plan.FieldID)
			}
			answer, err := c.askOther(otherPrompt)
			if err != nil {
				return "", err
			}
			return answer, nil
		}

		return selectedLabel, nil
	}
}

func (c *refinementConsole) askOther(prompt prompts.QuestionSpec) (string, error) {
	if _, err := fmt.Fprintf(c.out, "%s\n", prompt.Description); err != nil {
		return "", err
	}
	if placeholder := strings.TrimSpace(prompt.Placeholder); placeholder != "" {
		if _, err := fmt.Fprintf(c.out, "Hint: %s\n", placeholder); err != nil {
			return "", err
		}
	}

	for {
		answer, err := c.readLine("> ")
		if err != nil {
			return "", err
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			if _, err := fmt.Fprintf(c.errOut, "%s requires additional detail\n", prompt.FieldID); err != nil {
				return "", err
			}
			continue
		}
		return answer, nil
	}
}

func (c *refinementConsole) renderReview(report refinement.ValidationReport) {
	_, _ = fmt.Fprintln(c.out)
	_, _ = fmt.Fprintln(c.out, prompts.RenderReview(report))
	_, _ = fmt.Fprintln(c.out, "Commands: review | revise <field> | commit | quit")
}

func (c *refinementConsole) readLine(prefix string) (string, error) {
	if _, err := fmt.Fprint(c.out, prefix); err != nil {
		return "", err
	}

	line, err := c.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	line = strings.TrimRight(line, "\r\n")
	if err == io.EOF && strings.TrimSpace(line) == "" {
		return "", io.EOF
	}
	return line, nil
}

func resolveOption(options []prompts.PromptOption, raw string) (string, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}

	for index, option := range options {
		ordinal := fmt.Sprintf("%d", index+1)
		if raw == ordinal || strings.EqualFold(raw, option.Label) || strings.EqualFold(raw, option.Value) {
			return option.Value, option.Label, true
		}
	}

	return "", "", false
}

func findOtherPrompt(plan prompts.PromptPlan) (prompts.QuestionSpec, bool) {
	for _, prompt := range plan.Prompts {
		if prompt.Control == prompts.ControlTypeInput && strings.Contains(prompt.Key, "other") {
			return prompt, true
		}
	}
	return prompts.QuestionSpec{}, false
}

func flowReportOrFallback(state *refinement.SessionState, fallback refinement.ValidationReport) refinement.ValidationReport {
	report, err := refinement.DefaultValidator().Evaluate(state)
	if err != nil {
		return fallback
	}
	return report
}

func buildRefinementPayload(state *refinement.SessionState, result refinement.FlowResult) refinementPayload {
	snapshot := state.Snapshot()
	answers := make([]refinementPayloadAnswer, 0, len(snapshot))
	for _, field := range snapshot {
		answers = append(answers, refinementPayloadAnswer{
			FieldID:  field.Definition.ID,
			Section:  field.Definition.Section,
			Label:    field.Definition.Label,
			Value:    field.Answer.Value,
			Status:   field.Status,
			Revision: field.Answer.Revision,
		})
	}

	return refinementPayload{
		Phase:       "03-interactive-refinement-loop",
		State:       result.State,
		CommitReady: result.Report.CommitReady,
		Revision:    state.Revision(),
		Answers:     answers,
	}
}

func writePayload(out io.Writer, payload refinementPayload) error {
	if _, err := fmt.Fprintln(out, "\nCommitted refinement payload:"); err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "%s\n", encoded); err != nil {
		return err
	}
	return nil
}
