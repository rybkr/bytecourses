package middleware

import (
	"log"
	"net/http"

	"github.com/rybkr/bytecourses/internal/helpers"
	"github.com/rybkr/bytecourses/internal/models"
)

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return Auth(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			log.Println("user not found in context for admin check")
			helpers.Unauthorized(w, "unauthorized")
			return
		}

		if user.Role != models.RoleAdmin {
			log.Printf("user %s attempted to access admin endpoint without admin role", user.Email)
			helpers.Forbidden(w, "forbidden: admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}
