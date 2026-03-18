package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/hillscheck/internal/domain"
	"github.com/hillscheck/internal/infrastructure/auth"
)

type userIDKey struct{}
type userPlanKey struct{}

// NewAuth returns a middleware that validates JWT access tokens.
func NewAuth(jwt *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("Authorization")
			token := strings.TrimPrefix(raw, "Bearer ")
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := jwt.Validate(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey{}, claims.UserID)
			ctx = context.WithValue(ctx, userPlanKey{}, claims.Plan)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey{}).(string)
	return id
}

func PlanFromCtx(ctx context.Context) domain.Plan {
	plan, _ := ctx.Value(userPlanKey{}).(domain.Plan)
	return plan
}
