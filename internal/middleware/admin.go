package middleware

import (
	"github.com/rybkr/bytecourses/internal/models"
	"log"
	"net/http"
)

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return Auth(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			log.Println("user not found in context for admin check")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if user.Role != models.RoleAdmin {
			log.Printf("user %s attempted to access admin endpoint without admin role", user.Email)
			http.Error(w, "forbidden: admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
