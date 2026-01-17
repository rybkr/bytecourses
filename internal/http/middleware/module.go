package middleware

import (
	"bytecourses/internal/store"
	"net/http"
)

type ModuleIDFunc func(r *http.Request) (int64, bool)

func RequireModule(modules store.ModuleStore, moduleID ModuleIDFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mid, ok := moduleID(r)
			if !ok {
				http.Error(w, "missing module id", http.StatusBadRequest)
				return
			}

			m, ok := modules.GetModuleByID(r.Context(), mid)
			if !ok {
				http.Error(w, "module not found", http.StatusNotFound)
				return
			}

			next.ServeHTTP(w, r.WithContext(withModule(r.Context(), m)))
		})
	}
}
