package http

import (
	"net/http"

	"go.uber.org/zap"

	mw "github.com/hillscheck/internal/adapter/http/middleware"
	"github.com/hillscheck/internal/usecase"
)

type QuestHandler struct {
	getQuests *usecase.GetQuests
	log       *zap.Logger
}

func NewQuestHandler(getQuests *usecase.GetQuests, log *zap.Logger) *QuestHandler {
	return &QuestHandler{getQuests: getQuests, log: log}
}

func (h *QuestHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := mw.UserIDFromCtx(r.Context())
	if userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	quests, err := h.getQuests.Execute(r.Context(), userID)
	if err != nil {
		h.log.Error("get quests", zap.Error(err))
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonOK(w, http.StatusOK, quests)
}
