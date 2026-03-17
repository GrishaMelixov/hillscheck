package middleware

import (
	"context"
	"net/http"
	"strings"
)

type userIDKey struct{}

// Auth is a minimal JWT-stub middleware.
// In production this should validate the JWT and extract the user ID.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// TODO: validate JWT, extract real userID from claims.
		// For now we use the token value itself as a placeholder userID.
		ctx := context.WithValue(r.Context(), userIDKey{}, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey{}).(string)
	return id
}
