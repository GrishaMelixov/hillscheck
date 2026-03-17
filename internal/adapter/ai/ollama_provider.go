package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hillscheck/internal/usecase/port"
)

type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	return &OllamaProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (o *OllamaProvider) Classify(ctx context.Context, description string, mcc int) (port.ClassificationResult, error) {
	body, err := json.Marshal(ollamaRequest{
		Model:  o.model,
		Prompt: buildClassifyPrompt(description, mcc),
		Stream: false,
		Format: "json",
	})
	if err != nil {
		return port.ClassificationResult{}, fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return port.ClassificationResult{}, fmt.Errorf("build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return port.ClassificationResult{}, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return port.ClassificationResult{}, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var olResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&olResp); err != nil {
		return port.ClassificationResult{}, fmt.Errorf("decode ollama response: %w", err)
	}

	return parseClassifyJSON(olResp.Response)
}
