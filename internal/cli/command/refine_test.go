package command

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Nickbohm555/skill-cli/internal/refinement"
)

func TestRunRefineSessionCommitsStructuredPayload(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"Things maybe.",
		"Generate a Codex skill from one docs URL, including install steps, scope boundaries, and review-ready examples.",
		"Capture the docs source, extract implementation guidance, and turn it into a focused skill with explicit installation steps.",
		"The generated skill is installable, scoped to one domain, and includes concrete usage examples plus constraints.",
		"Use one docs URL only, keep the skill deterministic, and exclude unsupported setup steps or speculative workflows.",
		"Requires network access for docs fetches, a reachable documentation site, and OpenAI credentials only when structured summarization is enabled.",
		"Examples should include generating a skill from Go docs and refining boundaries when the docs mix tutorials with API references.",
		"Output examples must show install commands, supported inputs, and one explicit out-of-scope case for the final skill.",
		"In scope: extracting skill instructions, installation notes, supported commands, and concrete examples from the chosen docs set.",
		"Out of scope: building unrelated tooling, inventing missing APIs, or merging content from multiple unrelated documentation sites.",
		"commit",
		"",
	}, "\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	payload, err := runRefineSession(strings.NewReader(input), &stdout, &stderr)
	if err != nil {
		t.Fatalf("runRefineSession() error = %v", err)
	}

	if payload.State != refinement.FlowStateCommitted {
		t.Fatalf("payload.State = %q, want %q", payload.State, refinement.FlowStateCommitted)
	}
	if !payload.CommitReady {
		t.Fatal("payload.CommitReady = false, want true")
	}

	out := stdout.String()
	if !strings.Contains(out, "Summary before follow-up for purpose_summary") {
		t.Fatalf("stdout missing summarize-first handoff\n%s", out)
	}
	if !strings.Contains(out, "Committed refinement payload:") {
		t.Fatalf("stdout missing payload banner\n%s", out)
	}

	var decoded refinementPayload
	if err := decodePayloadJSON(out, &decoded); err != nil {
		t.Fatalf("decode payload json: %v", err)
	}
	if len(decoded.Answers) != len(refinement.DefaultFieldRegistry()) {
		t.Fatalf("decoded answers = %d, want %d", len(decoded.Answers), len(refinement.DefaultFieldRegistry()))
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRefineSessionBlocksCommitUntilRevisionImpactIsResolved(t *testing.T) {
	t.Parallel()

	input := strings.Join([]string{
		"Generate a Codex skill from one docs URL, including install steps, scope boundaries, and review-ready examples.",
		"Capture the docs source, extract implementation guidance, and turn it into a focused skill with explicit installation steps.",
		"The generated skill is installable, scoped to one domain, and includes concrete usage examples plus constraints.",
		"Use one docs URL only, keep the skill deterministic, and exclude unsupported setup steps or speculative workflows.",
		"Requires network access for docs fetches, a reachable documentation site, and OpenAI credentials only when structured summarization is enabled.",
		"Examples should include generating a skill from Go docs and refining boundaries when the docs mix tutorials with API references.",
		"Output examples must show install commands, supported inputs, and one explicit out-of-scope case for the final skill.",
		"In scope: extracting skill instructions, installation notes, supported commands, and concrete examples from the chosen docs set.",
		"Out of scope: building unrelated tooling, inventing missing APIs, or merging content from multiple unrelated documentation sites.",
		"revise purpose_summary",
		"Refocus the skill on Go documentation only, with deterministic extraction, install steps, and clear source boundaries.",
		"Extract the Go docs guidance from one source, turn it into a Codex skill, and keep the generated instructions installable and scoped.",
		"The skill installs cleanly, stays anchored to Go documentation, and includes concrete operating constraints plus usable examples.",
		"Examples should cover generating a Go docs skill and tightening scope when the source mixes reference material with tutorials.",
		"In scope: Go documentation extraction, skill instruction synthesis, install steps, and supported request examples from that source.",
		"Out of scope: mixing non-Go sources, inventing undocumented capabilities, or broadening the skill beyond the chosen Go docs site.",
		"commit",
		"revise example_outputs",
		"Output examples must show install commands, supported inputs, and one explicit out-of-scope case focused on Go docs only.",
		"commit",
		"",
	}, "\n")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	payload, err := runRefineSession(strings.NewReader(input), &stdout, &stderr)
	if err != nil {
		t.Fatalf("runRefineSession() error = %v", err)
	}

	if payload.State != refinement.FlowStateCommitted {
		t.Fatalf("payload.State = %q, want %q", payload.State, refinement.FlowStateCommitted)
	}

	if !strings.Contains(stderr.String(), "commit blocked: required fields are missing, unclear, or need revalidation") {
		t.Fatalf("stderr missing blocked-commit message\n%s", stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "Revising Purpose Summary (purpose_summary)") {
		t.Fatalf("stdout missing purpose revision prompt\n%s", out)
	}
	if !strings.Contains(out, "Revising Example Outputs (example_outputs)") {
		t.Fatalf("stdout missing example_outputs revision prompt\n%s", out)
	}
}

func decodePayloadJSON(out string, target *refinementPayload) error {
	start := strings.Index(out, "{")
	if start < 0 {
		return ioErr("json payload not found")
	}
	return json.Unmarshal([]byte(out[start:]), target)
}

type ioErr string

func (e ioErr) Error() string { return string(e) }
