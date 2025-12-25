package handlers

import (
	"bytecourses/internal/auth"
	"net/http"
)

type LogoutHandler struct {
	sessions auth.SessionStore
}

func NewLogoutHandler(sessions auth.SessionStore) *LogoutHandler {
	return &LogoutHandler{sessions: sessions}
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	c, err := r.Cookie("session")
	if err == nil {
		h.sessions.Delete(c.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusNoContent)
}
