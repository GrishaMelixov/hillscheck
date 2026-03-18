package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/GrishaMelixov/wealthcheck/internal/domain"
	"github.com/GrishaMelixov/wealthcheck/internal/infrastructure/auth"
	"github.com/GrishaMelixov/wealthcheck/internal/usecase"
)

type AuthHandler struct {
	register *usecase.RegisterUser
	login    *usecase.LoginUser
	refresh  *usecase.RefreshToken
	logout   *usecase.Logout
	log      *zap.Logger
}

func NewAuthHandler(
	register *usecase.RegisterUser,
	login *usecase.LoginUser,
	refresh *usecase.RefreshToken,
	logout *usecase.Logout,
	log *zap.Logger,
) *AuthHandler {
	return &AuthHandler{register: register, login: login, refresh: refresh, logout: logout, log: log}
}

// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Name == "" || body.Email == "" || len(body.Password) < 8 {
		http.Error(w, "name, email and password (min 8 chars) are required", http.StatusBadRequest)
		return
	}

	out, err := h.register.Execute(r.Context(), usecase.RegisterInput{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	})
	if errors.Is(err, domain.ErrEmailTaken) {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}
	if err != nil {
		h.log.Error("register", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":    out.User.ID,
		"name":  out.User.Name,
		"email": out.User.Email,
		"plan":  out.User.Plan,
	})
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	out, err := h.login.Execute(r.Context(), usecase.LoginInput{
		Email:    body.Email,
		Password: body.Password,
	})
	if errors.Is(err, domain.ErrWrongPassword) {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		h.log.Error("login", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	setRefreshCookie(w, out.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": out.AccessToken,
		"user": map[string]any{
			"id":    out.User.ID,
			"name":  out.User.Name,
			"email": out.User.Email,
			"plan":  out.User.Plan,
		},
	})
}

// POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	out, err := h.refresh.Execute(r.Context(), cookie.Value)
	if errors.Is(err, domain.ErrTokenInvalid) {
		http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}
	if err != nil {
		h.log.Error("refresh token", zap.Error(err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	setRefreshCookie(w, out.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]any{"access_token": out.AccessToken})
}

// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		_ = h.logout.Execute(r.Context(), cookie.Value)
	}
	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusNoContent)
}

func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // set true in production behind HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(auth.RefreshTokenTTL / time.Second),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
