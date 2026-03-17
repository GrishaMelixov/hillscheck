package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/hillscheck/internal/usecase/port"
)

type ReceiptUpload struct {
	uploadDir string
	pool      port.WorkerPool
	log       *zap.Logger
}

type ReceiptUploadResult struct {
	ReceiptID string
	FilePath  string
}

func NewReceiptUpload(uploadDir string, pool port.WorkerPool, log *zap.Logger) (*ReceiptUpload, error) {
	if err := os.MkdirAll(uploadDir, 0o750); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &ReceiptUpload{uploadDir: uploadDir, pool: pool, log: log}, nil
}

// Upload saves the file to disk and enqueues an OCR job.
func (u *ReceiptUpload) Upload(ctx context.Context, userID, filename string, r io.Reader) (ReceiptUploadResult, error) {
	receiptID := fmt.Sprintf("%d", time.Now().UnixNano())
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}

	safeFileName := receiptID + ext
	destPath := filepath.Join(u.uploadDir, userID)
	if err := os.MkdirAll(destPath, 0o750); err != nil {
		return ReceiptUploadResult{}, fmt.Errorf("create user upload dir: %w", err)
	}

	fullPath := filepath.Join(destPath, safeFileName)
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		return ReceiptUploadResult{}, fmt.Errorf("create receipt file: %w", err)
	}

	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(fullPath)
		return ReceiptUploadResult{}, fmt.Errorf("write receipt file: %w", err)
	}
	f.Close()

	result := ReceiptUploadResult{ReceiptID: receiptID, FilePath: fullPath}

	// Enqueue OCR processing (async).
	captured := result
	ocrJob := func(jCtx context.Context) error {
		u.log.Info("OCR job started", zap.String("receipt_id", captured.ReceiptID), zap.String("path", captured.FilePath))
		// OCR integration point — parse receipt, extract transaction data.
		return nil
	}

	if err := u.pool.Submit(ocrJob); err != nil {
		u.log.Warn("could not enqueue OCR job", zap.String("receipt_id", receiptID), zap.Error(err))
	}

	return result, nil
}
