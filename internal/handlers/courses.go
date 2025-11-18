package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "github.com/rybkr/bytecourses/internal/helpers"
    "github.com/rybkr/bytecourses/internal/middleware"
    "github.com/rybkr/bytecourses/internal/models"
    "github.com/rybkr/bytecourses/internal/store"
    "github.com/rybkr/bytecourses/internal/validation"
)

type CourseHandler struct {
    store *store.Store
}

func NewCourseHandler(store *store.Store) *CourseHandler {
    return &CourseHandler{store: store}
}

func (h *CourseHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        log.Println("user not found in context")
        helpers.Error(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    
    var course models.Course
    if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
        log.Printf("failed to decode course request: %v", err)
        helpers.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    v := validation.New()
    v.Required(course.Title, "title")
    v.MinLength(course.Title, 3, "title")
    v.MaxLength(course.Title, 255, "title")
    v.Required(course.Description, "description")
    v.MinLength(course.Description, 10, "description")
    
    if !v.Valid() {
        log.Printf("course validation failed: %v", v.Errors)
        helpers.JSON(w, http.StatusBadRequest, map[string]interface{}{
            "error": "validation failed",
            "fields": v.Errors,
        })
        return
    }
    
    course.InstructorID = user.ID
    course.Status = models.StatusPending
    
    if err := h.store.CreateCourse(r.Context(), &course); err != nil {
        log.Printf("failed to create course in handler: %v", err)
        helpers.Error(w, http.StatusInternalServerError, "internal server error")
        return
    }
    
    helpers.Created(w, course)
}

func (h *CourseHandler) ListCourses(w http.ResponseWriter, r *http.Request) {
    var status *models.CourseStatus
    if s := r.URL.Query().Get("status"); s != "" {
        st := models.CourseStatus(s)
        status = &st
    } else {
        approved := models.StatusApproved
        status = &approved
    }
    
    courses, err := h.store.GetCoursesWithInstructors(r.Context(), status)
    if err != nil {
        log.Printf("failed to get courses in handler: %v", err)
        helpers.Error(w, http.StatusInternalServerError, "internal server error")
        return
    }
    
    helpers.Success(w, courses)
}
