package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ParsedTransaction is the structured output from reading a bank screenshot.
type ParsedTransaction struct {
	Description string    `json:"description"`
	AmountCents int64     `json:"amount_cents"` // negative = expense, positive = income
	MCC         int       `json:"mcc"`
	OccurredAt  time.Time `json:"occurred_at"`
	Currency    string    `json:"currency"`
}

// GeminiVision parses bank screenshots using Gemini multimodal API.
type GeminiVision struct {
	client  *genai.Client
	model   string
	timeout time.Duration
}

func NewGeminiVision(apiKey, model string) (*GeminiVision, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("create gemini vision client: %w", err)
	}
	return &GeminiVision{
		client:  client,
		model:   model,
		timeout: 30 * time.Second,
	}, nil
}

func (g *GeminiVision) Close() {
	g.client.Close()
}

// ParseScreenshot sends an image to Gemini Vision and extracts transaction data.
// imageData is raw bytes; mimeType is "image/jpeg", "image/png", or "image/webp".
func (g *GeminiVision) ParseScreenshot(ctx context.Context, imageData []byte, mimeType string) ([]ParsedTransaction, error) {
	prompt := `You are a financial data extraction assistant. Analyze this bank app screenshot or receipt.

Extract ALL financial transactions visible. Return ONLY valid JSON, no markdown, no explanation.

JSON schema:
{
  "transactions": [
    {
      "description": "merchant or store name",
      "amount_rub": -1426.98,
      "mcc": 5816,
      "occurred_at": "2026-03-17T06:06:18",
      "currency": "RUB"
    }
  ]
}

Rules:
- amount_rub: negative for purchases/expenses, positive for income. Use the TOTAL amount for receipts.
- mcc: ISO 18245 code if visible in the screenshot, otherwise guess from merchant type (5812=restaurants, 5411=supermarkets, 5816=digital goods, 7941=sports, 8299=education, 5912=pharmacy). Use 0 if unknown.
- occurred_at: ISO 8601 format. If only date shown, use T12:00:00. Use current year if not shown.
- For кассовый чек (cash receipt): use the ИТОГО/total as one transaction, description = store name or first item.
- For bank transaction screens: description = merchant name shown.
- currency: always "RUB" unless explicitly shown otherwise.`

	tctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	imagePart := genai.ImageData(strings.TrimPrefix(mimeType, "image/"), imageData)
	resp, err := g.client.GenerativeModel(g.model).GenerateContent(tctx, imagePart, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini vision generate: %w", err)
	}

	raw := extractText(resp)
	return parseVisionJSON(raw)
}

type visionResponse struct {
	Transactions []struct {
		Description string  `json:"description"`
		AmountRub   float64 `json:"amount_rub"`
		MCC         int     `json:"mcc"`
		OccurredAt  string  `json:"occurred_at"`
		Currency    string  `json:"currency"`
	} `json:"transactions"`
}

func parseVisionJSON(raw string) ([]ParsedTransaction, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	// Find JSON object bounds
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	var r visionResponse
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, fmt.Errorf("parse vision response: %w (raw: %q)", err, raw)
	}

	results := make([]ParsedTransaction, 0, len(r.Transactions))
	for _, t := range r.Transactions {
		cur := t.Currency
		if cur == "" {
			cur = "RUB"
		}
		// Parse date
		occurredAt := time.Now()
		for _, layout := range []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02",
		} {
			if parsed, err := time.Parse(layout, t.OccurredAt); err == nil {
				occurredAt = parsed
				break
			}
		}
		results = append(results, ParsedTransaction{
			Description: t.Description,
			AmountCents: int64(t.AmountRub * 100),
			MCC:         t.MCC,
			OccurredAt:  occurredAt,
			Currency:    cur,
		})
	}
	return results, nil
}
