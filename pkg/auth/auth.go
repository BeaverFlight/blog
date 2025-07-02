package auth

import (
	"blog/pkg/models"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

func GenerateJWT(login string, secret string, expiration time.Duration) (string, error) {
	claims := &models.Claims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Генерация JWT токена
// swagger:response jwtToken
type JWTResponse struct {
	// in:body
	Body struct {
		Token string `json:"token"`
	}
}

func AuthMiddleware(secret string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(rw, "Authorization header required", http.StatusUnauthorized)
				return
			}
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(rw, "Invalid token format", http.StatusUnauthorized)
				return
			}

			tokenString := tokenParts[1]
			claims := &models.Claims{}

			token, err := jwt.ParseWithClaims(
				tokenString,
				claims,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(secret), nil
				},
			)
			log.Println(token)

			if err != nil || !token.Valid {
				http.Error(rw, "Invalid token", http.StatusUnauthorized)
				return
			}
			// Добавляем логин в контекст
			ctx := context.WithValue(r.Context(), "login", claims.Login)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

// Middleware аутентификации
// swagger:parameters protectedOperation
type AuthHeader struct {
	// Bearer токен
	// in: header
	// required: true
	Authorization string
}
