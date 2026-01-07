package middleware

import (
	"bytecourses/internal/store"
	"net/http"
)

// ProposalIDFunc extracts a proposal ID from the request.
// This decouples proposal middleware from any specific router.
type ProposalIDFunc func(r *http.Request) (int64, bool)

// RequireProposal loads a proposal by ID and injects it into the request context.
// It returns 400 on missing ID and 404 if the proposal does not exist.
func RequireProposal(proposals store.ProposalStore, proposalID ProposalIDFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pid, ok := proposalID(r)
			if !ok {
				http.Error(w, "missing id", http.StatusBadRequest)
				return
			}

			p, ok := proposals.GetProposalByID(r.Context(), pid)
			if !ok {
				http.Error(w, "proposal not found", http.StatusNotFound)
				return
			}

			next.ServeHTTP(w, r.WithContext(withProposal(r.Context(), p)))
		})
	}
}

// RequireProposalOwner enforces that the authenticated user owns the proposal.
// It must be used after RequireUser and RequireProposal.
func RequireProposalOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		p, ok := ProposalFromContext(r.Context())
		if !ok {
			http.Error(w, "proposal not found", http.StatusNotFound)
			return
		}
		if p.AuthorID != u.ID {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
