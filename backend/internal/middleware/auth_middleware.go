package middleware

import (
	"net/http"

	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/auth"
	"github.com/vasantharatnam/taskflow-sabbavarapu/backend/internal/response"
)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			tokenString, err := auth.ExtractBearerToken(header)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, err := auth.ParseToken(jwtSecret, tokenString)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			user := auth.AuthenticatedUser{
				UserID: claims.UserID,
				Email:  claims.Email,
			}

			ctx := auth.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}