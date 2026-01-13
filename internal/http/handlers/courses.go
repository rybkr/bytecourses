package handlers

import (
	"bytecourses/internal/services"
	"net/http"
)

type CourseHandler struct {
	services *services.Services
}

func NewCourseHandler(services *services.Services) *CourseHandler {
	return &CourseHandler{
		services: services,
	}
}

func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var request services.CreateCourseRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.InstructorID = user.ID

	course, err := h.services.Courses.CreateCourse(r.Context(), &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, course)
}

func (h *CourseHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	c, ok := requireCourse(w, r)
	if !ok {
		return
	}

	course, err := h.services.Courses.GetCourse(r.Context(), c, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, course)
}
