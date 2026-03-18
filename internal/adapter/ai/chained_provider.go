package ai

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/GrishaMelixov/wealthcheck/internal/usecase/port"
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
	// ── Слой 1: детерминированная MCC-таблица (50+ категорий, точность 100%) ──
	// Проверяем MCC первым. Если код известен — возвращаем сразу, LLM не нужен.
	// Это и есть "гибридность": дорогой AI-вызов делается только для неизвестных MCC.
	if entry, ok := LookupMCC(mcc); ok {
		c.log.Debug("mcc table hit", zap.Int("mcc", mcc), zap.String("category", entry.Category))
		return mccEntryToResult(entry), nil
	}

	// ── Слой 2: LLM провайдеры (Gemini → Ollama) ─────────────────────────────
	var lastErr error
	for _, p := range c.providers {
		result, err := p.Classify(ctx, description, mcc)
		if err == nil {
			return result, nil
		}
		c.log.Warn("provider classify failed, trying next", zap.Error(err))
		lastErr = err
	}

	// ── Слой 3: статический fallback — pipeline никогда не блокируется ────────
	c.log.Warn("all providers failed, using static fallback", zap.Error(lastErr))
	return staticFallback(mcc), nil
}

// mccEntryToResult конвертирует запись из MCC-таблицы в ClassificationResult.
// Выбирает доминирующий RPG-атрибут (с наибольшим abs-значением).
func mccEntryToResult(e MCCEntry) port.ClassificationResult {
	type kv struct {
		attr string
		val  int
	}
	candidates := []kv{
		{port.AttrHP, e.HP},
		{port.AttrStrength, e.Strength},
		{port.AttrIntellect, e.Intellect},
		{port.AttrMana, e.Mana},
		{port.AttrLuck, e.Luck},
		{port.AttrXP, e.XP},
	}
	best := kv{port.AttrXP, e.XP}
	for _, c := range candidates {
		if abs(c.val) > abs(best.val) {
			best = c
		}
	}
	if best.val == 0 {
		best = kv{port.AttrXP, 1}
	}
	return result(e.Category, best.attr, best.val)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// staticFallback — последний рубеж, когда ни MCC-таблица, ни LLM не сработали.
func staticFallback(mcc int) port.ClassificationResult {
	_ = mcc // MCC уже был проверен в таблице выше — здесь он не нужен
	return result(DefaultEntry.Category, port.AttrXP, DefaultEntry.XP)
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
