package ai

import "context"

// VisionParser is the interface for image-to-transaction providers.
// Implementations: TesseractVision (free, local), GeminiVision (cloud).
type VisionParser interface {
	ParseScreenshot(ctx context.Context, imageData []byte, mimeType string) ([]ParsedTransaction, error)
}
