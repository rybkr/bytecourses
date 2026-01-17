package middleware

import (
	"bytecourses/internal/store"
	"net/http"
)

type ContentItemIDFunc func(r *http.Request) (int64, bool)

func RequireContentItem(content store.ContentStore, contentItemID ContentItemIDFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cid, ok := contentItemID(r)
			if !ok {
				http.Error(w, "missing content id", http.StatusBadRequest)
				return
			}

			item, ok := content.GetContentItemByID(r.Context(), cid)
			if !ok {
				http.Error(w, "content not found", http.StatusNotFound)
				return
			}

			next.ServeHTTP(w, r.WithContext(withContentItem(r.Context(), item)))
		})
	}
}
