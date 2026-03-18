package http

import (
	"net/http"
	"strconv"

	"go.uber.org/zap"

	mw "github.com/GrishaMelixov/wealthcheck/internal/adapter/http/middleware"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase"
)


type AnalyticsHandler struct {
	getAnalytics *usecase.GetAnalytics
	log          *zap.Logger
}

func NewAnalyticsHandler(getAnalytics *usecase.GetAnalytics, log *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{getAnalytics: getAnalytics, log: log}
}

// Summary godoc
// GET /api/v1/analytics/summary?account_id=<uuid>&days=30
//
// Возвращает агрегированную аналитику за период.
// Первый вызов — PostgreSQL (100-200ms), повторные — Redis (< 1ms).
// TTL кеша: 1 час. Инвалидируется автоматически после импорта транзакций.
func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	userID := mw.UserIDFromCtx(r.Context())
	if userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accountID := r.URL.Query().Get("account_id")
	if accountID == "" || accountID == "undefined" || accountID == "null" {
		jsonError(w, "account_id is required", http.StatusBadRequest)
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 365 {
			days = v
		}
	}

	summary, err := h.getAnalytics.Execute(r.Context(), accountID, days)
	if err != nil {
		h.log.Error("analytics summary", zap.String("user_id", userID), zap.Error(err))
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonOK(w, http.StatusOK, summary)
}
