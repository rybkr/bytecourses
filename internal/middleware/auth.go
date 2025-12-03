package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rybkr/bytecourses/internal/helpers"
	"github.com/rybkr/bytecourses/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}
	return []byte(secret)
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("missing authorization header")
			helpers.Unauthorized(w, "unauthorized")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Println("invalid authorization header format")
			helpers.Unauthorized(w, "unauthorized")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return getJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			log.Printf("invalid token: %v", err)
			helpers.Unauthorized(w, "unauthorized")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("invalid token claims")
			helpers.Unauthorized(w, "unauthorized")
			return
		}

		userID := int(claims["user_id"].(float64))
		email := claims["email"].(string)
		role := models.UserRole(claims["role"].(string))

		user := &models.User{
			ID:    userID,
			Email: email,
			Role:  role,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

func OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, jwt.ErrSignatureInvalid
					}
					return getJWTSecret(), nil
				})

				if err == nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						userID := int(claims["user_id"].(float64))
						email := claims["email"].(string)
						role := models.UserRole(claims["role"].(string))

						user := &models.User{
							ID:    userID,
							Email: email,
							Role:  role,
						}

						ctx := context.WithValue(r.Context(), UserContextKey, user)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	}
}
