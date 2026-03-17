package http

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	mw "github.com/hillscheck/internal/adapter/http/middleware"
	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/usecase"
)

type ProfileHandler struct {
	getProfile *usecase.GetProfile
	log        *zap.Logger
}

func NewProfileHandler(getProfile *usecase.GetProfile, log *zap.Logger) *ProfileHandler {
	return &ProfileHandler{getProfile: getProfile, log: log}
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := mw.UserIDFromCtx(r.Context())
	if userID == "" {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := h.getProfile.Execute(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			jsonError(w, "profile not found", http.StatusNotFound)
			return
		}
		h.log.Error("get profile", zap.Error(err))
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	jsonOK(w, http.StatusOK, profile)
}
