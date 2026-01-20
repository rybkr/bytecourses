package handlers

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/infrastructure/persistence/memory"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/services"
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func setupProposalService() *services.ProposalService {
	proposalRepo := memory.NewProposalRepository()
	userRepo := memory.NewUserRepository()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	eventBus := events.NewInMemoryEventBus(logger)

	return services.NewProposalService(proposalRepo, userRepo, eventBus)
}

func TestProposalHandler_Create(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	body := `{"title":"Test Proposal","summary":"Summary","qualifications":"Quals","target_audience":"Audience","learning_objectives":"Objectives","outline":"Outline","assumed_prerequisites":"Prerequisites"}`
	req := httptest.NewRequest("POST", "/proposals", bytes.NewReader([]byte(body))).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create: expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestProposalHandler_Update(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create a proposal first
	proposal, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Original Title",
		Summary:  "Original Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}

	body := `{"title":"Updated Title"}`
	req := requestWithChiContextAndUser("/proposals/"+strconv.FormatInt(proposal.ID, 10), map[string]string{"id": strconv.FormatInt(proposal.ID, 10)}, ctx)
	req.Method = "PUT"
	req.Body = http.MaxBytesReader(nil, io.NopCloser(bytes.NewReader([]byte(body))), 1<<20)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Update: expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

func TestProposalHandler_Submit(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create a proposal first
	_, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Test Proposal",
		Summary:  "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}

	req := requestWithChiContextAndUser("/proposals/1/submit", map[string]string{"id": "1"}, ctx)
	req.Method = "POST"
	w := httptest.NewRecorder()

	handler.Submit(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Submit: expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

func TestProposalHandler_Withdraw(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create and submit a proposal first
	_, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Test Proposal",
		Summary:  "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}
	_, err = proposalService.Submit(context.Background(), &services.SubmitProposalInput{
		ProposalID: 1,
		UserID:     user.ID,
	})
	if err != nil {
		t.Fatalf("Setup: failed to submit proposal: %v", err)
	}

	req := requestWithChiContextAndUser("/proposals/1/withdraw", map[string]string{"id": "1"}, ctx)
	req.Method = "POST"
	w := httptest.NewRecorder()

	handler.Withdraw(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Withdraw: expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

func TestProposalHandler_Review(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	adminUser := &domain.User{ID: 2, Role: domain.UserRoleAdmin}
	adminCtx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), adminUser)

	// Create and submit a proposal first
	_, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Test Proposal",
		Summary:  "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}
	_, err = proposalService.Submit(context.Background(), &services.SubmitProposalInput{
		ProposalID: 1,
		UserID:     user.ID,
	})
	if err != nil {
		t.Fatalf("Setup: failed to submit proposal: %v", err)
	}

	body := `{"decision":"approve","notes":"Looks good"}`
	req := requestWithChiContextAndUser("/proposals/1/review", map[string]string{"id": "1"}, adminCtx)
	req.Method = "POST"
	req.Body = http.MaxBytesReader(nil, io.NopCloser(bytes.NewReader([]byte(body))), 1<<20)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Review(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Review: expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestProposalHandler_Delete(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create a proposal first
	_, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Test Proposal",
		Summary:  "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}

	req := requestWithChiContextAndUser("/proposals/1", map[string]string{"id": "1"}, ctx)
	req.Method = "DELETE"
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Delete: expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

func TestProposalHandler_GetByID(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create a proposal first
	_, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Test Proposal",
		Summary:  "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}

	req := requestWithChiContextAndUser("/proposals/1", map[string]string{"id": "1"}, ctx)
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetByID: expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestProposalHandler_ListAll(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	adminUser := &domain.User{ID: 1, Role: domain.UserRoleAdmin}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), adminUser)

	req := httptest.NewRequest("GET", "/proposals", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ListAll(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ListAll: expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestProposalHandler_ListMine(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	req := httptest.NewRequest("GET", "/proposals/mine", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ListMine(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ListMine: expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestProposalHandler_Approve(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	adminUser := &domain.User{ID: 2, Role: domain.UserRoleAdmin}
	adminCtx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), adminUser)

	// Create and submit a proposal first
	_, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Test Proposal",
		Summary:  "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}
	_, err = proposalService.Submit(context.Background(), &services.SubmitProposalInput{
		ProposalID: 1,
		UserID:     user.ID,
	})
	if err != nil {
		t.Fatalf("Setup: failed to submit proposal: %v", err)
	}

	body := `{"notes":"Approved"}`
	req := requestWithChiContextAndUser("/proposals/1/approve", map[string]string{"id": "1"}, adminCtx)
	req.Method = "POST"
	req.Body = io.NopCloser(bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Approve(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Approve: expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

func TestProposalHandler_Reject(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	adminUser := &domain.User{ID: 2, Role: domain.UserRoleAdmin}
	adminCtx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), adminUser)

	// Create and submit a proposal first
	_, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Test Proposal",
		Summary:  "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}
	_, err = proposalService.Submit(context.Background(), &services.SubmitProposalInput{
		ProposalID: 1,
		UserID:     user.ID,
	})
	if err != nil {
		t.Fatalf("Setup: failed to submit proposal: %v", err)
	}

	body := `{"notes":"Rejected"}`
	req := requestWithChiContextAndUser("/proposals/1/reject", map[string]string{"id": "1"}, adminCtx)
	req.Method = "POST"
	req.Body = io.NopCloser(bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Reject(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Reject: expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

func TestProposalHandler_RequestChanges(t *testing.T) {
	proposalService := setupProposalService()
	handler := NewProposalHandler(proposalService)

	user := &domain.User{ID: 1}
	adminUser := &domain.User{ID: 2, Role: domain.UserRoleAdmin}
	adminCtx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), adminUser)

	// Create and submit a proposal first
	_, err := proposalService.Create(context.Background(), &services.CreateProposalInput{
		AuthorID: user.ID,
		Title:    "Test Proposal",
		Summary:  "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create proposal: %v", err)
	}
	_, err = proposalService.Submit(context.Background(), &services.SubmitProposalInput{
		ProposalID: 1,
		UserID:     user.ID,
	})
	if err != nil {
		t.Fatalf("Setup: failed to submit proposal: %v", err)
	}

	body := `{"notes":"Needs changes"}`
	req := requestWithChiContextAndUser("/proposals/1/request-changes", map[string]string{"id": "1"}, adminCtx)
	req.Method = "POST"
	req.Body = io.NopCloser(bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.RequestChanges(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("RequestChanges: expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}
