package auth

import (
	"blog/pkg/models"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

var (
	secret     string
	expiration time.Duration
)

func InitializationSecret(newSecret string, hour int) {
	secret = newSecret
	expiration = time.Duration(hour) * time.Hour
}

func GenerateJWT(login string) (string, error) {
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

// swagger:response jwtToken
type JWTResponse struct {
	// in:body
	Body struct {
		Token string `json:"token"`
	}
}

func AuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				models.ResponseBadRequest(rw)
				return
			}
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				models.ResponseBadRequest(rw)
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

			if err != nil || !token.Valid {
				models.ResponseUnauthorized(rw)
				return
			}
			// Добавляем логин в контекст
			ctx := context.WithValue(r.Context(), "login", claims.Login)
			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

// swagger:parameters createArticle updateArticle deleteArticle
type AuthHeader struct {
	// Bearer токен
	// in: header
	// required: true
	Authorization string
}
