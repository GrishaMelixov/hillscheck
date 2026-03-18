package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	mw "github.com/hillscheck/internal/adapter/http/middleware"
	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/usecase"
	"github.com/hillscheck/internal/usecase/port"
)

type TransactionHandler struct {
	importer *usecase.TransactionImport
	txRepo   port.TransactionRepository
	accounts port.AccountRepository
	log      *zap.Logger
}

func NewTransactionHandler(
	importer *usecase.TransactionImport,
	txRepo port.TransactionRepository,
	accounts port.AccountRepository,
	log *zap.Logger,
) *TransactionHandler {
	return &TransactionHandler{importer: importer, txRepo: txRepo, accounts: accounts, log: log}
}

type importTxInput struct {
	ExternalID          string `json:"external_id"`
	Amount              int64  `json:"amount"`
	MCC                 int    `json:"mcc"`
	OriginalDescription string `json:"original_description"`
	OccurredAt          string `json:"occurred_at"` // RFC3339
}

type importRequest struct {
	AccountID    string          `json:"account_id"`
	Transactions []importTxInput `json:"transactions"`
}

func (h *TransactionHandler) Import(w http.ResponseWriter, r *http.Request) {
	var req importRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Transactions) == 0 {
		jsonError(w, "transactions array must not be empty", http.StatusBadRequest)
		return
	}

	params := make([]port.CreateTransactionParams, 0, len(req.Transactions))
	for _, t := range req.Transactions {
		if t.ExternalID == "" {
			jsonError(w, "each transaction must have external_id", http.StatusBadRequest)
			return
		}
		if t.Amount == 0 {
			jsonError(w, "amount must be non-zero", http.StatusBadRequest)
			return
		}
		occurredAt := time.Now()
		if t.OccurredAt != "" {
			var err error
			occurredAt, err = time.Parse(time.RFC3339, t.OccurredAt)
			if err != nil {
				jsonError(w, "occurred_at must be RFC3339", http.StatusBadRequest)
				return
			}
		}
		params = append(params, port.CreateTransactionParams{
			ExternalID:          t.ExternalID,
			Amount:              t.Amount,
			MCC:                 t.MCC,
			OriginalDescription: t.OriginalDescription,
			OccurredAt:          occurredAt,
		})
	}

	result, err := h.importer.Import(r.Context(), usecase.ImportRequest{
		AccountID:    req.AccountID,
		Transactions: params,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.log.Error("import transactions", zap.Error(err))
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonOK(w, http.StatusAccepted, map[string]any{
		"created":    len(result.Created),
		"duplicates": len(result.Duplicates),
	})
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")
	if accountID == "" {
		jsonError(w, "account_id query param is required", http.StatusBadRequest)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	txs, err := h.txRepo.ListByAccount(r.Context(), accountID, limit, offset)
	if err != nil {
		h.log.Error("list transactions", zap.Error(err))
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = mw.GetRequestID(r.Context())
	jsonOK(w, http.StatusOK, txs)
}

func (h *TransactionHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID := mw.UserIDFromCtx(r.Context())
	if userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accounts, err := h.accounts.ListByUser(r.Context(), userID)
	if err != nil {
		h.log.Error("list accounts", zap.Error(err))
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonOK(w, http.StatusOK, map[string]any{"accounts": accounts})
}

func jsonOK(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
