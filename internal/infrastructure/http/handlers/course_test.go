package handlers

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/infrastructure/persistence/memory"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/services"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
)

func setupCourseService() *services.CourseService {
	courseRepo := memory.NewCourseRepository()
	proposalRepo := memory.NewProposalRepository()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	eventBus := events.NewInMemoryEventBus(logger)

	return services.NewCourseService(courseRepo, proposalRepo, eventBus)
}

func requestWithChiContext(path string, routeParams map[string]string) *http.Request {
	r := httptest.NewRequest("GET", path, nil)
	rctx := chi.NewRouteContext()
	for key, val := range routeParams {
		rctx.URLParams.Add(key, val)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func requestWithChiContextAndUser(path string, routeParams map[string]string, userCtx context.Context) *http.Request {
	r := httptest.NewRequest("GET", path, nil)
	rctx := chi.NewRouteContext()
	for key, val := range routeParams {
		rctx.URLParams.Add(key, val)
	}
	// Merge user context with chi route context
	return r.WithContext(context.WithValue(userCtx, chi.RouteCtxKey, rctx))
}

func TestCourseHandler_Create(t *testing.T) {
	courseService := setupCourseService()
	handler := NewCourseHandler(courseService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	body := `{"title":"Test Course","summary":"Summary","target_audience":"Audience","learning_objectives":"Objectives","assumed_prerequisites":"Prerequisites"}`
	req := httptest.NewRequest("POST", "/courses", bytes.NewReader([]byte(body))).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create: expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestCourseHandler_Update(t *testing.T) {
	courseService := setupCourseService()
	handler := NewCourseHandler(courseService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create a course first
	_, err := courseService.Create(context.Background(), &services.CreateCourseInput{
		InstructorID: user.ID,
		Title:        "Original Title",
		Summary:      "Original Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create course: %v", err)
	}

	t.Run("ValidUpdate", func(t *testing.T) {
		body := `{"title":"Updated Title","summary":"Updated Summary"}`
		req := requestWithChiContextAndUser("/courses/1", map[string]string{"id": "1"}, ctx)
		req.Method = "PUT"
		req.Body = http.MaxBytesReader(nil, io.NopCloser(bytes.NewReader([]byte(body))), 1<<20)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Update(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("Update: expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
		}
	})

	t.Run("InvalidID", func(t *testing.T) {
		req := requestWithChiContextAndUser("/courses/invalid", map[string]string{"id": "invalid"}, ctx)
		req.Method = "PUT"
		w := httptest.NewRecorder()

		handler.Update(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Update: expected status %d for invalid ID, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestCourseHandler_Publish(t *testing.T) {
	courseService := setupCourseService()
	handler := NewCourseHandler(courseService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create a course first
	_, err := courseService.Create(context.Background(), &services.CreateCourseInput{
		InstructorID: user.ID,
		Title:        "Test Course",
		Summary:      "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create course: %v", err)
	}

	req := requestWithChiContextAndUser("/courses/1/publish", map[string]string{"id": "1"}, ctx)
	req.Method = "POST"
	w := httptest.NewRecorder()

	handler.Publish(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Publish: expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestCourseHandler_CreateFromProposal(t *testing.T) {
	// For this test we need both proposal and course services
	proposalRepo := memory.NewProposalRepository()
	userRepo := memory.NewUserRepository()
	courseRepo := memory.NewCourseRepository()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	eventBus := events.NewInMemoryEventBus(logger)

	proposalService := services.NewProposalService(proposalRepo, userRepo, eventBus)
	courseService := services.NewCourseService(courseRepo, proposalRepo, eventBus)
	handler := NewCourseHandler(courseService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create and approve a proposal first
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
	_, err = proposalService.Review(context.Background(), &services.ReviewProposalInput{
		ProposalID: 1,
		ReviewerID: 2,
		Decision:   services.ReviewDecisionApprove,
	})
	if err != nil {
		t.Fatalf("Setup: failed to approve proposal: %v", err)
	}

	body := `{"proposal_id":1}`
	req := httptest.NewRequest("POST", "/courses/from-proposal", bytes.NewReader([]byte(body))).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateFromProposal(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("CreateFromProposal: expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}

func TestCourseHandler_GetByID(t *testing.T) {
	courseService := setupCourseService()
	handler := NewCourseHandler(courseService)

	user := &domain.User{ID: 1}
	ctx := middleware.WithUser(middleware.WithSession(context.Background(), "session123"), user)

	// Create a course first
	_, err := courseService.Create(context.Background(), &services.CreateCourseInput{
		InstructorID: user.ID,
		Title:        "Test Course",
		Summary:      "Test Summary",
	})
	if err != nil {
		t.Fatalf("Setup: failed to create course: %v", err)
	}

	req := requestWithChiContextAndUser("/courses/1", map[string]string{"id": "1"}, ctx)
	w := httptest.NewRecorder()

	handler.GetByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetByID: expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var result domain.Course
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("GetByID: failed to decode response: %v", err)
	}
	if result.ID != 1 {
		t.Fatalf("GetByID: expected course ID 1, got %d", result.ID)
	}
}

func TestCourseHandler_ListLive(t *testing.T) {
	t.Run("ValidList", func(t *testing.T) {
		courseService := setupCourseService()
		handler := NewCourseHandler(courseService)

		user := &domain.User{ID: 1}

		// Create and publish courses
		_, err := courseService.Create(context.Background(), &services.CreateCourseInput{
			InstructorID: user.ID,
			Title:        "Course 1",
			Summary:      "Summary 1",
		})
		if err != nil {
			t.Fatalf("Setup: failed to create course: %v", err)
		}
		_, err = courseService.Publish(context.Background(), &services.PublishCourseInput{
			CourseID: 1,
			UserID:   user.ID,
		})
		if err != nil {
			t.Fatalf("Setup: failed to publish course: %v", err)
		}

		req := httptest.NewRequest("GET", "/courses/live", nil)
		w := httptest.NewRecorder()

		handler.ListLive(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ListLive: expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result []domain.Course
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("ListLive: failed to decode response: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("ListLive: expected 1 course, got %d", len(result))
		}
	})

	t.Run("EmptyList", func(t *testing.T) {
		courseService := setupCourseService()
		handler := NewCourseHandler(courseService)

		req := httptest.NewRequest("GET", "/courses/live", nil)
		w := httptest.NewRecorder()

		handler.ListLive(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("ListLive: expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result []domain.Course
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("ListLive: failed to decode response: %v", err)
		}
		if result == nil {
			t.Fatalf("ListLive: expected empty slice, got nil")
		}
		if len(result) != 0 {
			t.Fatalf("ListLive: expected empty slice, got %d items", len(result))
		}
	})
}

func intPtr(i int64) *int64 {
	return &i
}
