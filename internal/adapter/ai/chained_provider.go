package ai

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/hillscheck/internal/usecase/port"
)

// ChainedProvider tries each provider in order and returns the first success.
// If all fail, it falls back to a static MCC-based rule so the pipeline never blocks.
type ChainedProvider struct {
	providers []port.CategoryProvider
	log       *zap.Logger
}

func NewChainedProvider(log *zap.Logger, providers ...port.CategoryProvider) *ChainedProvider {
	return &ChainedProvider{providers: providers, log: log}
}

func (c *ChainedProvider) Classify(ctx context.Context, description string, mcc int) (port.ClassificationResult, error) {
	var lastErr error
	for _, p := range c.providers {
		result, err := p.Classify(ctx, description, mcc)
		if err == nil {
			return result, nil
		}
		c.log.Warn("provider classify failed, trying next", zap.Error(err))
		lastErr = err
	}

	// Static fallback — always succeeds, never blocks the pipeline.
	c.log.Warn("all providers failed, using static fallback", zap.Error(lastErr))
	return staticFallback(mcc), nil
}

// staticFallback maps common MCC ranges to deterministic RPG impacts.
func staticFallback(mcc int) port.ClassificationResult {
	switch {
	case mcc >= 5940 && mcc <= 5949: // sporting goods
		return result("Sports & Fitness", port.AttrStrength, 3)
	case mcc >= 5940 && mcc <= 5945: // hobby/toy
		return result("Hobby", port.AttrMana, 3)
	case mcc == 5942: // book stores
		return result("Books & Learning", port.AttrIntellect, 5)
	case mcc >= 5811 && mcc <= 5814: // restaurants / fast food
		return result("Dining", port.AttrHP, -2)
	case mcc >= 8000 && mcc <= 8099: // health / medical
		return result("Healthcare", port.AttrHP, 5)
	case mcc >= 7993 && mcc <= 7999: // amusement / gambling
		return result("Entertainment", port.AttrMana, 2)
	case mcc >= 5200 && mcc <= 5211: // hardware / home improvement
		return result("Home Improvement", port.AttrStrength, 2)
	default:
		return result("General Purchase", port.AttrXP, 1)
	}
}

func result(category, attribute string, value int) port.ClassificationResult {
	return port.ClassificationResult{
		Category: category,
		Impact:   port.Impact{Attribute: attribute, Value: value},
	}
}

// NewProviderFromConfig builds the appropriate provider based on the config strategy.
func NewProviderFromConfig(
	strategy string,
	geminiKey, geminiModel string,
	ollamaURL, ollamaModel string,
	log *zap.Logger,
) (port.CategoryProvider, error) {
	switch strategy {
	case "gemini":
		return NewGeminiProvider(geminiKey, geminiModel)
	case "ollama":
		return NewOllamaProvider(ollamaURL, ollamaModel), nil
	case "chained":
		var providers []port.CategoryProvider
		if geminiKey != "" {
			g, err := NewGeminiProvider(geminiKey, geminiModel)
			if err != nil {
				log.Warn("failed to init gemini, skipping", zap.Error(err))
			} else {
				providers = append(providers, g)
			}
		}
		providers = append(providers, NewOllamaProvider(ollamaURL, ollamaModel))
		return NewChainedProvider(log, providers...), nil
	default:
		return nil, fmt.Errorf("unknown provider strategy: %q", strategy)
	}
}
