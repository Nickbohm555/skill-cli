package content

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

const (
	defaultSummaryModel     = openai.ChatModelGPT5Mini
	defaultSummaryMaxTokens = 180
	defaultSummaryMaxRunes  = 240
	summaryConfidenceHigh   = "high"
	summaryConfidenceMedium = "medium"
	summaryConfidenceLow    = "low"
	fallbackSummaryNote     = "deterministic fallback summary"
	openAIKeyEnv            = "OPENAI_API_KEY"
)

// ChunkSummary is the stable, attribution-linked summary record produced for a
// chunk and consumed by later review rendering.
type ChunkSummary struct {
	ChunkID        string
	SourceURL      string
	Summary        string
	Confidence     string
	Notes          string
	Attribution    ChunkAttribution
	UsedFallback   bool
	FallbackReason string
}

// SummaryConfig controls chunk summarization behavior.
type SummaryConfig struct {
	Provider SummaryProvider
	Model    string
}

// SummaryProvider abstracts the structured summary backend so later tests can
// inject failures without depending on a live API call.
type SummaryProvider interface {
	Summarize(ctx context.Context, input SummaryInput) (SummaryRecord, error)
}

// SummaryInput is the provider-facing chunk payload used to preserve stable
// identifiers and attribution during summarization.
type SummaryInput struct {
	Text        string
	Attribution ChunkAttribution
}

// SummaryRecord is the schema-shaped summary payload returned by providers
// before it is validated and normalized for downstream use.
type SummaryRecord struct {
	ChunkID    string `json:"chunk_id"`
	SourceURL  string `json:"source_url"`
	Summary    string `json:"summary"`
	Confidence string `json:"confidence,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

// OpenAISummaryProvider uses the Responses API structured output path to
// request a bounded JSON summary record.
type OpenAISummaryProvider struct {
	Client openai.Client
	Model  string
}

// SummarizeChunks produces one structured summary per attributed chunk,
// falling back deterministically when the provider is unavailable or fails.
func SummarizeChunks(ctx context.Context, chunks []AttributedChunk) ([]ChunkSummary, error) {
	return SummarizeChunksWithConfig(ctx, chunks, SummaryConfig{})
}

// SummarizeChunksWithConfig allows callers to override the provider/model while
// preserving the same local validation and fallback behavior.
func SummarizeChunksWithConfig(ctx context.Context, chunks []AttributedChunk, cfg SummaryConfig) ([]ChunkSummary, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	cfg = normalizeSummaryConfig(cfg)
	summaries := make([]ChunkSummary, 0, len(chunks))

	for _, chunk := range chunks {
		input, err := newSummaryInputRecord(chunk)
		if err != nil {
			return nil, err
		}

		record, usedFallback, fallbackReason := summarizeOne(ctx, cfg.Provider, input)
		summary, err := finalizeSummaryRecord(record, input, usedFallback, fallbackReason)
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// NewDefaultSummaryProvider builds the preferred provider chain for this
// repository: OpenAI structured output when credentials are available, else nil
// so deterministic fallback mode is used.
func NewDefaultSummaryProvider(model string) SummaryProvider {
	if strings.TrimSpace(os.Getenv(openAIKeyEnv)) == "" {
		return nil
	}

	return OpenAISummaryProvider{
		Client: openai.NewClient(),
		Model:  normalizeSummaryModel(model),
	}
}

func (p OpenAISummaryProvider) Summarize(ctx context.Context, input SummaryInput) (SummaryRecord, error) {
	payload := buildSummaryPrompt(input)
	resp, err := p.Client.Responses.New(ctx, responses.ResponseNewParams{
		Model: normalizeSummaryModel(p.Model),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(payload),
		},
		MaxOutputTokens: openai.Int(defaultSummaryMaxTokens),
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigParamOfJSONSchema(
				"chunk_summary",
				chunkSummarySchema(),
			),
		},
	})
	if err != nil {
		return SummaryRecord{}, err
	}

	output := strings.TrimSpace(resp.OutputText())
	if output == "" {
		return SummaryRecord{}, fmt.Errorf("structured summary response was empty")
	}

	var record SummaryRecord
	if err := json.Unmarshal([]byte(output), &record); err != nil {
		return SummaryRecord{}, fmt.Errorf("decode structured summary: %w", err)
	}

	return record, nil
}

func summarizeOne(ctx context.Context, provider SummaryProvider, input SummaryInput) (SummaryRecord, bool, string) {
	if provider == nil {
		return fallbackSummaryRecord(input, ""), true, "provider unavailable"
	}

	record, err := provider.Summarize(ctx, input)
	if err != nil {
		return fallbackSummaryRecord(input, ""), true, err.Error()
	}

	return record, false, ""
}

func finalizeSummaryRecord(record SummaryRecord, input SummaryInput, usedFallback bool, fallbackReason string) (ChunkSummary, error) {
	if usedFallback {
		record = fallbackSummaryRecord(input, fallbackReason)
	}

	normalized, err := normalizeSummaryRecord(record, input)
	if err != nil {
		record = fallbackSummaryRecord(input, err.Error())
		normalized, err = normalizeSummaryRecord(record, input)
		if err != nil {
			return ChunkSummary{}, err
		}
		usedFallback = true
		if fallbackReason == "" {
			fallbackReason = "summary validation failed"
		}
	}

	summary := ChunkSummary{
		ChunkID:        normalized.ChunkID,
		SourceURL:      normalized.SourceURL,
		Summary:        normalized.Summary,
		Confidence:     normalized.Confidence,
		Notes:          normalized.Notes,
		Attribution:    cloneAttribution(input.Attribution),
		UsedFallback:   usedFallback,
		FallbackReason: strings.TrimSpace(fallbackReason),
	}
	if summary.UsedFallback && summary.Notes == "" {
		summary.Notes = fallbackSummaryNote
	}

	return summary, nil
}

func normalizeSummaryConfig(cfg SummaryConfig) SummaryConfig {
	if cfg.Provider == nil {
		cfg.Provider = NewDefaultSummaryProvider(cfg.Model)
	}
	cfg.Model = normalizeSummaryModel(cfg.Model)
	return cfg
}

func normalizeSummaryModel(model string) string {
	if strings.TrimSpace(model) == "" {
		return defaultSummaryModel
	}
	return strings.TrimSpace(model)
}

func newSummaryInputRecord(chunk AttributedChunk) (SummaryInput, error) {
	if !chunk.Attribution.HasRequiredFields() {
		return SummaryInput{}, fmt.Errorf("summary input missing attribution fields for chunk %q", chunk.Chunk.ID)
	}

	text := strings.TrimSpace(chunk.Chunk.Text)
	if text == "" {
		return SummaryInput{}, fmt.Errorf("summary input has empty chunk text for chunk %q", chunk.Chunk.ID)
	}

	return SummaryInput{
		Text:        text,
		Attribution: cloneAttribution(chunk.Attribution),
	}, nil
}

func normalizeSummaryRecord(record SummaryRecord, input SummaryInput) (SummaryRecord, error) {
	record.ChunkID = strings.TrimSpace(record.ChunkID)
	record.SourceURL = strings.TrimSpace(record.SourceURL)
	record.Summary = normalizeSummaryText(record.Summary)
	record.Confidence = normalizeSummaryConfidence(record.Confidence)
	record.Notes = trimRunes(strings.TrimSpace(record.Notes), defaultSummaryMaxRunes/2)

	if record.ChunkID == "" {
		return SummaryRecord{}, fmt.Errorf("summary record missing chunk_id")
	}
	if record.SourceURL == "" {
		return SummaryRecord{}, fmt.Errorf("summary record missing source_url")
	}
	if record.ChunkID != input.Attribution.ChunkID {
		return SummaryRecord{}, fmt.Errorf("summary record chunk_id %q did not match input %q", record.ChunkID, input.Attribution.ChunkID)
	}
	if record.SourceURL != input.Attribution.SourceURL {
		return SummaryRecord{}, fmt.Errorf("summary record source_url %q did not match input %q", record.SourceURL, input.Attribution.SourceURL)
	}
	if record.Summary == "" {
		return SummaryRecord{}, fmt.Errorf("summary record missing summary text")
	}
	if countSummaryLines(record.Summary) > 2 {
		return SummaryRecord{}, fmt.Errorf("summary record exceeded 2 lines")
	}

	return record, nil
}

func fallbackSummaryRecord(input SummaryInput, reason string) SummaryRecord {
	summary := fallbackSummaryText(input)
	notes := fallbackSummaryNote
	if trimmed := strings.TrimSpace(reason); trimmed != "" {
		notes = notes + ": " + trimmed
	}

	return SummaryRecord{
		ChunkID:    input.Attribution.ChunkID,
		SourceURL:  input.Attribution.SourceURL,
		Summary:    summary,
		Confidence: summaryConfidenceLow,
		Notes:      trimRunes(notes, defaultSummaryMaxRunes/2),
	}
}

func fallbackSummaryText(input SummaryInput) string {
	lines := make([]string, 0, 2)

	contextLine := buildFallbackContextLine(input.Attribution)
	if contextLine != "" {
		lines = append(lines, contextLine)
	}

	bodyLine := buildFallbackBodyLine(input.Text)
	if bodyLine != "" {
		if len(lines) == 0 {
			lines = append(lines, bodyLine)
		} else if !strings.EqualFold(lines[0], bodyLine) {
			lines = append(lines, bodyLine)
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "Chunk content available for review.")
	}

	if len(lines) > 2 {
		lines = lines[:2]
	}

	return strings.Join(lines, "\n")
}

func buildFallbackContextLine(attr ChunkAttribution) string {
	label := firstNonEmpty(
		lastHeading(attr.HeadingPath),
		attr.PageTitle,
		attr.SourceURL,
	)
	if label == "" {
		return ""
	}

	return trimRunes(cleanSummarySentence(label), defaultSummaryMaxRunes/2)
}

func buildFallbackBodyLine(text string) string {
	clean := cleanSummarySourceText(text)
	if clean == "" {
		return ""
	}

	sentences := splitSummarySentences(clean)
	for _, sentence := range sentences {
		candidate := trimRunes(cleanSummarySentence(sentence), defaultSummaryMaxRunes/2)
		if candidate != "" {
			return candidate
		}
	}

	return trimRunes(cleanSummarySentence(clean), defaultSummaryMaxRunes/2)
}

func buildSummaryPrompt(input SummaryInput) string {
	var builder strings.Builder
	builder.WriteString("Summarize this documentation chunk as JSON.\n")
	builder.WriteString("Return exactly these fields: chunk_id, source_url, summary, confidence, notes.\n")
	builder.WriteString("Rules:\n")
	builder.WriteString("- summary must be 1 or 2 concise lines total.\n")
	builder.WriteString("- chunk_id and source_url must exactly match the provided values.\n")
	builder.WriteString("- confidence, if present, must be one of high, medium, low.\n")
	builder.WriteString("- notes is optional and should stay brief.\n\n")
	builder.WriteString("chunk_id: ")
	builder.WriteString(input.Attribution.ChunkID)
	builder.WriteString("\nsource_url: ")
	builder.WriteString(input.Attribution.SourceURL)
	builder.WriteString("\npage_title: ")
	builder.WriteString(input.Attribution.PageTitle)
	builder.WriteString("\nheading_path: ")
	builder.WriteString(strings.Join(input.Attribution.HeadingPath, " > "))
	builder.WriteString("\nreference: ")
	builder.WriteString(input.Attribution.Reference)
	builder.WriteString("\n\nchunk_text:\n")
	builder.WriteString(input.Text)

	return builder.String()
}

func chunkSummarySchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"chunk_id", "source_url", "summary"},
		"properties": map[string]any{
			"chunk_id": map[string]any{
				"type":      "string",
				"minLength": 1,
			},
			"source_url": map[string]any{
				"type":      "string",
				"minLength": 1,
			},
			"summary": map[string]any{
				"type":      "string",
				"minLength": 1,
				"maxLength": defaultSummaryMaxRunes,
			},
			"confidence": map[string]any{
				"type": "string",
				"enum": []string{summaryConfidenceHigh, summaryConfidenceMedium, summaryConfidenceLow},
			},
			"notes": map[string]any{
				"type":      "string",
				"maxLength": defaultSummaryMaxRunes / 2,
			},
		},
	}
}

func normalizeSummaryText(input string) string {
	parts := strings.Split(strings.ReplaceAll(strings.TrimSpace(input), "\r\n", "\n"), "\n")
	lines := make([]string, 0, 2)
	for _, part := range parts {
		cleaned := cleanSummarySentence(part)
		if cleaned == "" {
			continue
		}
		lines = append(lines, trimRunes(cleaned, defaultSummaryMaxRunes/2))
		if len(lines) == 2 {
			break
		}
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n")
}

func normalizeSummaryConfidence(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case summaryConfidenceHigh:
		return summaryConfidenceHigh
	case summaryConfidenceLow:
		return summaryConfidenceLow
	case "", summaryConfidenceMedium:
		return summaryConfidenceMedium
	default:
		return summaryConfidenceMedium
	}
}

func countSummaryLines(input string) int {
	count := 0
	for _, part := range strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(part) != "" {
			count++
		}
	}
	return count
}

func cleanSummarySourceText(input string) string {
	replacer := strings.NewReplacer(
		"\r\n", "\n",
		"\r", "\n",
		"`", "",
		"*", "",
		"_", "",
		"#", "",
		">", "",
		"|", " ",
		"[", "",
		"]", "",
		"(", " ",
		")", " ",
	)

	lines := strings.Split(replacer.Replace(input), "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := cleanSummarySentence(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(trimmed), "http://") || strings.HasPrefix(strings.ToLower(trimmed), "https://") {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}

	return strings.Join(cleaned, " ")
}

func cleanSummarySentence(input string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(input)), " ")
}

func splitSummarySentences(input string) []string {
	input = cleanSummarySentence(input)
	if input == "" {
		return nil
	}

	parts := make([]string, 0, 4)
	start := 0
	for i, r := range input {
		if r != '.' && r != '!' && r != '?' {
			continue
		}

		next := i + utf8.RuneLen(r)
		if next < len(input) && input[next] != ' ' {
			continue
		}

		part := cleanSummarySentence(input[start:next])
		if part != "" {
			parts = append(parts, part)
		}
		start = next
	}

	if tail := cleanSummarySentence(input[start:]); tail != "" {
		parts = append(parts, tail)
	}

	if len(parts) == 0 {
		return []string{input}
	}

	return parts
}

func trimRunes(input string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(input) <= limit {
		return input
	}

	runes := []rune(input)
	if limit <= 1 {
		return string(runes[:limit])
	}

	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}

func lastHeading(path []string) string {
	if len(path) == 0 {
		return ""
	}
	return strings.TrimSpace(path[len(path)-1])
}
