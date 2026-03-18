package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
)

type GeminiProvider struct {
	client  *genai.Client
	model   string
	timeout time.Duration
}

func NewGeminiProvider(apiKey, model string) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}
	return &GeminiProvider{
		client:  client,
		model:   model,
		timeout: 15 * time.Second,
	}, nil
}

func (g *GeminiProvider) Classify(ctx context.Context, description string, mcc int) (port.ClassificationResult, error) {
	prompt := buildClassifyPrompt(description, mcc)

	tctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	resp, err := g.client.GenerativeModel(g.model).GenerateContent(tctx, genai.Text(prompt))
	if err != nil {
		return port.ClassificationResult{}, fmt.Errorf("gemini generate: %w", err)
	}

	raw := extractText(resp)
	return parseClassifyJSON(raw)
}

func (g *GeminiProvider) Close() {
	g.client.Close()
}

func buildClassifyPrompt(description string, mcc int) string {
	return fmt.Sprintf(`Classify this financial transaction. Return ONLY valid JSON, no markdown, no extra text.
JSON schema: {"category": string, "impact": {"attribute": string, "value": int}}
Allowed attributes: xp, hp, mana, strength, intellect, luck.
Rules:
- Books, education, courses → intellect +5..+15, xp +10..+30
- Gym, sports → strength +5..+10, hp +5..+10
- Food (healthy) → hp +3..+8
- Fast food, junk → hp -3..-8
- Entertainment, games → mana +5..+10
- Medical → hp +5..+15
- Gambling → luck -5..-15
- Default → xp +1..+5

Transaction: "%s"
MCC: %d`, description, mcc)
}

func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	c := resp.Candidates[0]
	if c.Content == nil || len(c.Content.Parts) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, p := range c.Content.Parts {
		if t, ok := p.(genai.Text); ok {
			sb.WriteString(string(t))
		}
	}
	return sb.String()
}

type classifyResponse struct {
	Category string `json:"category"`
	Impact   struct {
		Attribute string `json:"attribute"`
		Value     int    `json:"value"`
	} `json:"impact"`
}

func parseClassifyJSON(raw string) (port.ClassificationResult, error) {
	// Strip potential markdown code fences
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var r classifyResponse
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return port.ClassificationResult{}, fmt.Errorf("parse classify response: %w (raw: %q)", err, raw)
	}
	return port.ClassificationResult{
		Category: r.Category,
		Impact: port.Impact{
			Attribute: r.Impact.Attribute,
			Value:     r.Impact.Value,
		},
	}, nil
}
