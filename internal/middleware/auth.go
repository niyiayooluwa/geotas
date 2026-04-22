package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/niyiayooluwa/geotas/internal/model"
)

// a custom type for context keys
// avoids collisions with other packages using the same string key
type contextKey string

const UserIDKey contextKey = "user_id"
const UserEmailKey contextKey = "email"

func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// read the Authorization header
		var authHeader string = r.Header.Get("Authorization")

		// if empty, block the request
		if authHeader == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		// must start with "Bearer"
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// strip "Bearer " prefix to get raw token string
		var tokenString string = strings.TrimPrefix(authHeader, "Bearer ")

		// parse and validate the token
		token, err := jwt.ParseWithClaims(
			tokenString,
			&model.Claims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			},
		)

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// extract claims from the validated token
		claims, ok := token.Claims.(*model.Claims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		//stuff user_id and email into request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		// pass request to next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
