package http

import (
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/hillscheck/internal/adapter/ai"
	mw "github.com/hillscheck/internal/adapter/http/middleware"
	"github.com/hillscheck/internal/usecase"
)

type ReceiptHandler struct {
	uploader *usecase.ReceiptUpload
	vision   ai.VisionParser
	log      *zap.Logger
}

func NewReceiptHandler(uploader *usecase.ReceiptUpload, vision ai.VisionParser, log *zap.Logger) *ReceiptHandler {
	return &ReceiptHandler{uploader: uploader, vision: vision, log: log}
}

func (h *ReceiptHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := mw.UserIDFromCtx(r.Context())
	if userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	const maxSize = 10 << 20 // 10 MB
	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonError(w, "request too large or not multipart", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	result, err := h.uploader.Upload(r.Context(), userID, header.Filename, file)
	if err != nil {
		h.log.Error("upload receipt", zap.Error(err))
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonOK(w, http.StatusAccepted, map[string]string{
		"receipt_id": result.ReceiptID,
	})
}

// Parse accepts an image (JPEG/PNG/WEBP) and uses Gemini Vision to extract transactions.
// Returns a list of parsed transactions the client can review before importing.
func (h *ReceiptHandler) Parse(w http.ResponseWriter, r *http.Request) {
	_ = mw.UserIDFromCtx(r.Context()) // auth already enforced by middleware

	const maxSize = 10 << 20 // 10 MB
	if err := r.ParseMultipartForm(maxSize); err != nil {
		jsonError(w, "request too large or not multipart", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Detect MIME type from extension or content-type header.
	mimeType := detectMIME(header.Filename, header.Header.Get("Content-Type"))
	if mimeType == "" {
		jsonError(w, "unsupported file type: use JPEG, PNG, or WEBP", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		jsonError(w, "failed to read file", http.StatusBadRequest)
		return
	}

	txs, err := h.vision.ParseScreenshot(r.Context(), data, mimeType)
	if err != nil {
		h.log.Error("gemini vision parse", zap.Error(err))
		errMsg := "Не удалось распознать скриншот"
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "Quota") {
			errMsg = "Превышена квота Gemini API — включи billing на ai.google.dev"
		}
		jsonError(w, errMsg, http.StatusUnprocessableEntity)
		return
	}

	if len(txs) == 0 {
		jsonError(w, "no transactions found in screenshot", http.StatusUnprocessableEntity)
		return
	}

	// Convert to wire format the frontend ImportModal expects.
	type wireTransaction struct {
		Description string `json:"description"`
		AmountCents int64  `json:"amount_cents"`
		MCC         int    `json:"mcc"`
		OccurredAt  string `json:"occurred_at"`
		Currency    string `json:"currency"`
	}
	out := make([]wireTransaction, 0, len(txs))
	for _, t := range txs {
		out = append(out, wireTransaction{
			Description: t.Description,
			AmountCents: t.AmountCents,
			MCC:         t.MCC,
			OccurredAt:  t.OccurredAt.Format("2006-01-02T15:04:05"),
			Currency:    t.Currency,
		})
	}

	jsonOK(w, http.StatusOK, map[string]any{"transactions": out})
}

func detectMIME(filename, contentType string) string {
	ext := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") ||
		strings.Contains(contentType, "jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(ext, ".png") || strings.Contains(contentType, "png"):
		return "image/png"
	case strings.HasSuffix(ext, ".webp") || strings.Contains(contentType, "webp"):
		return "image/webp"
	case strings.HasSuffix(ext, ".heic") || strings.HasSuffix(ext, ".heif"):
		return "image/heic"
	}
	return ""
}
