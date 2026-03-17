package http

import (
	"net/http"

	"go.uber.org/zap"

	mw "github.com/hillscheck/internal/adapter/http/middleware"
	"github.com/hillscheck/internal/usecase"
)

type ReceiptHandler struct {
	uploader *usecase.ReceiptUpload
	log      *zap.Logger
}

func NewReceiptHandler(uploader *usecase.ReceiptUpload, log *zap.Logger) *ReceiptHandler {
	return &ReceiptHandler{uploader: uploader, log: log}
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
