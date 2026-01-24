package handlers

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/services"
)

type PageData struct {
	User *domain.User
	Data any
}

type ProposalPageData struct {
	User             *domain.User
	Proposal         *domain.Proposal
	CourseExists     bool
	ExistingCourseID *int64
}

type CoursesPageData struct {
	User         *domain.User
	Courses      []domain.Course
	Instructors  map[int64]*domain.User
	ModuleCounts map[int64]int
}

type CoursePageData struct {
	User             *domain.User
	Course           *domain.Course
	Instructor       *domain.User
	IsInstructor     bool
	Modules          []domain.Module
	ReadingsByModule map[int64][]domain.Reading
	ActiveNavItem    string
}

type CourseEditPageData struct {
	User             *domain.User
	Course           *domain.Course
	Instructor       *domain.User
	IsInstructor     bool
	Modules          []domain.Module
	ReadingsByModule map[int64][]domain.Reading
	ActiveNavItem    string
}

type CourseContentPageData struct {
	User             *domain.User
	Course           *domain.Course
	Instructor       *domain.User
	IsInstructor     bool
	Modules          []domain.Module
	ReadingsByModule map[int64][]domain.Reading
	CurrentReading   *domain.Reading
	CurrentModule    *domain.Module
	PreviousReading  *domain.Reading
	NextReading      *domain.Reading
	ActiveNavItem    string
}

type PageHandler struct {
	templates       map[string]*template.Template
	funcMap         template.FuncMap
	proposalService *services.ProposalService
	courseService   *services.CourseService
	moduleService   *services.ModuleService
	contentService  *services.ContentService
	userRepo        persistence.UserRepository
}

func NewPageHandler(templatesFS embed.FS, proposalService *services.ProposalService, courseService *services.CourseService, moduleService *services.ModuleService, contentService *services.ContentService, userRepo persistence.UserRepository) *PageHandler {
	funcMap := template.FuncMap{
		"markdown": renderMarkdown,
	}

	h := &PageHandler{
		templates:       make(map[string]*template.Template),
		funcMap:         funcMap,
		proposalService: proposalService,
		courseService:   courseService,
		moduleService:   moduleService,
		contentService:  contentService,
		userRepo:        userRepo,
	}

	layoutContent, err := fs.ReadFile(templatesFS, "templates/layout.html")
	if err != nil {
		return nil
	}

	partialsContent := ""
	if partialEntries, err := fs.ReadDir(templatesFS, "templates/partials"); err == nil {
		for _, e := range partialEntries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".html") {
				if content, err := fs.ReadFile(templatesFS, "templates/partials/"+e.Name()); err == nil {
					partialsContent += string(content)
				}
			}
		}
	}

	entries, err := fs.ReadDir(templatesFS, "templates/pages")
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		pageContent, err := fs.ReadFile(templatesFS, "templates/pages/"+entry.Name())
		if err != nil {
			return nil
		}

		tmpl, err := template.New("").Funcs(funcMap).Parse(string(layoutContent))
		if err != nil {
			return nil
		}

		if partialsContent != "" {
			tmpl, err = tmpl.Parse(partialsContent)
			if err != nil {
				return nil
			}
		}

		tmpl, err = tmpl.Parse(string(pageContent))
		if err != nil {
			return nil
		}

		h.templates[entry.Name()] = tmpl
	}

	return h
}

func renderMarkdown(s string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(s), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(s))
	}
	return template.HTML(buf.String())
}

func (h *PageHandler) render(w http.ResponseWriter, r *http.Request, name string, data any) {
	tmpl, ok := h.templates[name]
	if !ok {
		log.Printf("template not found: %s", name)
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	user, _ := middleware.UserFromContext(r.Context())

	pd := PageData{
		User: user,
		Data: data,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	h.render(w, r, "404.html", nil)
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "home.html", nil)
}

func (h *PageHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "login.html", nil)
}

func (h *PageHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "register.html", nil)
}

func (h *PageHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "forgot_password.html", nil)
}

func (h *PageHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "reset_password.html", nil)
}

func (h *PageHandler) Profile(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "profile.html", nil)
}

func (h *PageHandler) Courses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courseService.List(r.Context())
	if err != nil {
		log.Printf("error fetching courses: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	instructors := make(map[int64]*domain.User)
	instructorIDs := make(map[int64]bool)
	for _, course := range courses {
		if !instructorIDs[course.InstructorID] {
			instructorIDs[course.InstructorID] = true
			if instructor, ok := h.userRepo.GetByID(r.Context(), course.InstructorID); ok {
				instructors[course.InstructorID] = instructor
			}
		}
	}

	user, _ := middleware.UserFromContext(r.Context())

	moduleCounts := make(map[int64]int)
	if user != nil {
		var userID int64
		var userRole domain.UserRole
		userID = user.ID
		userRole = user.Role
		for _, course := range courses {
			modules, err := h.moduleService.List(r.Context(), &services.ListModulesQuery{
				CourseID: course.ID,
				UserID:   userID,
				UserRole: userRole,
			})
			if err == nil {
				moduleCounts[course.ID] = len(modules)
			}
		}
	}

	pd := CoursesPageData{
		User:         user,
		Courses:      courses,
		Instructors:  instructors,
		ModuleCounts: moduleCounts,
	}

	tmpl, ok := h.templates["courses.html"]
	if !ok {
		log.Printf("template not found: courses.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) CourseView(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var userID int64
	var userRole domain.UserRole
	if user != nil {
		userID = user.ID
		userRole = user.Role
	}

	course, err := h.courseService.Get(r.Context(), &services.GetCourseQuery{
		CourseID: courseID,
		UserID:   userID,
		UserRole: userRole,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching course: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var instructor *domain.User
	if inst, ok := h.userRepo.GetByID(r.Context(), course.InstructorID); ok {
		instructor = inst
	}

	isInstructor := user != nil && course.IsTaughtBy(user)

	var modules []domain.Module
	readingsByModule := make(map[int64][]domain.Reading)

	if user != nil {
		modulesList, err := h.moduleService.List(r.Context(), &services.ListModulesQuery{
			CourseID: courseID,
			UserID:   userID,
			UserRole: userRole,
		})
		if err == nil {
			modules = modulesList
			for _, module := range modules {
				items, err := h.contentService.List(r.Context(), &services.ListContentQuery{
					ModuleID: module.ID,
					UserID:   userID,
					UserRole: userRole,
				})
				if err == nil {
					readings := make([]domain.Reading, 0, len(items))
					for _, item := range items {
						if reading, ok := item.(*domain.Reading); ok {
							readings = append(readings, *reading)
						}
					}
					readingsByModule[module.ID] = readings
				}
			}
		}
	}

	pd := CoursePageData{
		User:             user,
		Course:           course,
		Instructor:       instructor,
		IsInstructor:     isInstructor,
		Modules:          modules,
		ReadingsByModule: readingsByModule,
		ActiveNavItem:    "home",
	}

	tmpl, ok := h.templates["course_view.html"]
	if !ok {
		log.Printf("template not found: course_view.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) CourseEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	course, err := h.courseService.Get(r.Context(), &services.GetCourseQuery{
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching course: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !course.IsTaughtBy(user) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var instructor *domain.User
	if inst, ok := h.userRepo.GetByID(r.Context(), course.InstructorID); ok {
		instructor = inst
	}

	modulesList, err := h.moduleService.List(r.Context(), &services.ListModulesQuery{
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		log.Printf("error fetching modules: %v", err)
		modulesList = []domain.Module{}
	}

	readingsByModule := make(map[int64][]domain.Reading)
	for _, module := range modulesList {
		items, err := h.contentService.List(r.Context(), &services.ListContentQuery{
			ModuleID: module.ID,
			UserID:   user.ID,
			UserRole: user.Role,
		})
		if err == nil {
			readings := make([]domain.Reading, 0, len(items))
			for _, item := range items {
				if reading, ok := item.(*domain.Reading); ok {
					readings = append(readings, *reading)
				}
			}
			readingsByModule[module.ID] = readings
		}
	}

	pd := CourseEditPageData{
		User:             user,
		Course:           course,
		Instructor:       instructor,
		IsInstructor:     true,
		Modules:          modulesList,
		ReadingsByModule: readingsByModule,
		ActiveNavItem:    "settings",
	}

	tmpl, ok := h.templates["course_edit.html"]
	if !ok {
		log.Printf("template not found: course_edit.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) CourseContent(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var userID int64
	var userRole domain.UserRole
	if user != nil {
		userID = user.ID
		userRole = user.Role
	}

	course, err := h.courseService.Get(r.Context(), &services.GetCourseQuery{
		CourseID: courseID,
		UserID:   userID,
		UserRole: userRole,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching course: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var instructor *domain.User
	if inst, ok := h.userRepo.GetByID(r.Context(), course.InstructorID); ok {
		instructor = inst
	}

	isInstructor := user != nil && course.IsTaughtBy(user)

	modulesList, err := h.moduleService.List(r.Context(), &services.ListModulesQuery{
		CourseID: courseID,
		UserID:   userID,
		UserRole: userRole,
	})
	if err != nil {
		log.Printf("error fetching modules: %v", err)
		modulesList = []domain.Module{}
	}

	readingsByModule := make(map[int64][]domain.Reading)
	var allReadings []struct {
		Module  domain.Module
		Reading domain.Reading
	}

	for i := range modulesList {
		module := &modulesList[i]
		items, err := h.contentService.List(r.Context(), &services.ListContentQuery{
			ModuleID: module.ID,
			UserID:   userID,
			UserRole: userRole,
		})
		if err == nil {
			readings := make([]domain.Reading, 0, len(items))
			for _, item := range items {
				if reading, ok := item.(*domain.Reading); ok {
					readings = append(readings, *reading)
				}
			}
			readingsByModule[module.ID] = readings
			for j := range readings {
				allReadings = append(allReadings, struct {
					Module  domain.Module
					Reading domain.Reading
				}{Module: *module, Reading: readings[j]})
			}
		}
	}

	var currentModule *domain.Module
	var currentReading *domain.Reading
	var previousReading *domain.Reading
	var nextReading *domain.Reading

	readingIDStr := r.URL.Query().Get("readingId")
	if readingIDStr != "" {
		readingID, err := strconv.ParseInt(readingIDStr, 10, 64)
		if err == nil {
			for i, item := range allReadings {
				if item.Reading.ID == readingID {
					currentModule = &item.Module
					currentReading = &item.Reading
					if i > 0 {
						prev := allReadings[i-1]
						previousReading = &prev.Reading
					}
					if i < len(allReadings)-1 {
						next := allReadings[i+1]
						nextReading = &next.Reading
					}
					break
				}
			}
		}
	}

	if currentReading == nil && len(allReadings) > 0 {
		first := allReadings[0]
		currentModule = &first.Module
		currentReading = &first.Reading
		if len(allReadings) > 1 {
			second := allReadings[1]
			nextReading = &second.Reading
		}
	}

	pd := CourseContentPageData{
		User:             user,
		Course:           course,
		Instructor:       instructor,
		IsInstructor:     isInstructor,
		Modules:          modulesList,
		ReadingsByModule: readingsByModule,
		CurrentReading:   currentReading,
		CurrentModule:    currentModule,
		PreviousReading:  previousReading,
		NextReading:      nextReading,
		ActiveNavItem:    "content",
	}

	tmpl, ok := h.templates["course_content.html"]
	if !ok {
		log.Printf("template not found: course_content.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) Proposals(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "proposals.html", nil)
}

func (h *PageHandler) ProposalView(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	proposalID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	proposal, err := h.proposalService.Get(r.Context(), &services.GetProposalQuery{
		ProposalID: proposalID,
		UserID:     user.ID,
		UserRole:   user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "proposal not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching proposal: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, ok := h.templates["proposal_view.html"]
	if !ok {
		log.Printf("template not found: proposal_view.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	pd := ProposalPageData{
		User:         user,
		Proposal:     proposal,
		CourseExists: false,
	}

	if proposal.Status == domain.ProposalStatusApproved && proposal.AuthorID == user.ID {
		existing, ok := h.courseService.Courses.GetByProposalID(r.Context(), proposalID)
		if ok && existing != nil {
			pd.CourseExists = true
			pd.ExistingCourseID = &existing.ID
		}
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) ProposalEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	proposalIDStr := chi.URLParam(r, "id")
	if proposalIDStr == "" {
		tmpl, ok := h.templates["proposal_edit.html"]
		if !ok {
			log.Printf("template not found: proposal_edit.html")
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}

		pd := ProposalPageData{
			User: user,
			Proposal: &domain.Proposal{
				ID:     0,
				Status: domain.ProposalStatusDraft,
			},
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
			log.Printf("template execution error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		buf.WriteTo(w)
		return
	}

	proposalID, err := strconv.ParseInt(proposalIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	proposal, err := h.proposalService.Get(r.Context(), &services.GetProposalQuery{
		ProposalID: proposalID,
		UserID:     user.ID,
		UserRole:   user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "proposal not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching proposal: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, ok := h.templates["proposal_edit.html"]
	if !ok {
		log.Printf("template not found: proposal_edit.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	pd := ProposalPageData{
		User:     user,
		Proposal: proposal,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

type ReadingPageData struct {
	User         *domain.User
	Course       *domain.Course
	Module       *domain.Module
	Reading      *domain.Reading
	IsInstructor bool
}

type ContentNewPageData struct {
	User          *domain.User
	Course        *domain.Course
	Module        *domain.Module
	IsInstructor  bool
	ActiveNavItem string
}

func (h *PageHandler) LectureView(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	readingID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	content, err := h.contentService.Get(r.Context(), &services.GetContentQuery{
		ContentID: readingID,
		ModuleID:  moduleID,
		UserID:    user.ID,
		UserRole:  user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "reading not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching reading: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	reading, ok := content.(*domain.Reading)
	if !ok {
		http.Error(w, "invalid content type", http.StatusInternalServerError)
		return
	}

	course, err := h.courseService.Get(r.Context(), &services.GetCourseQuery{
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		http.Error(w, "course not found", http.StatusNotFound)
		return
	}

	module, err := h.moduleService.Get(r.Context(), &services.GetModuleQuery{
		ModuleID: moduleID,
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		http.Error(w, "module not found", http.StatusNotFound)
		return
	}

	isInstructor := course.IsTaughtBy(user)

	pd := ReadingPageData{
		User:         user,
		Course:       course,
		Module:       module,
		Reading:      reading,
		IsInstructor: isInstructor,
	}

	tmpl, ok := h.templates["lecture_view.html"]
	if !ok {
		log.Printf("template not found: lecture_view.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) LectureEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	readingID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	content, err := h.contentService.Get(r.Context(), &services.GetContentQuery{
		ContentID: readingID,
		ModuleID:  moduleID,
		UserID:    user.ID,
		UserRole:  user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "reading not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching reading: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	reading, ok := content.(*domain.Reading)
	if !ok {
		http.Error(w, "invalid content type", http.StatusInternalServerError)
		return
	}

	course, err := h.courseService.Get(r.Context(), &services.GetCourseQuery{
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		http.Error(w, "course not found", http.StatusNotFound)
		return
	}

	module, err := h.moduleService.Get(r.Context(), &services.GetModuleQuery{
		ModuleID: moduleID,
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		http.Error(w, "module not found", http.StatusNotFound)
		return
	}

	isInstructor := course.IsTaughtBy(user)

	pd := ReadingPageData{
		User:         user,
		Course:       course,
		Module:       module,
		Reading:      reading,
		IsInstructor: isInstructor,
	}

	tmpl, ok := h.templates["lecture_edit.html"]
	if !ok {
		log.Printf("template not found: lecture_edit.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) ContentNew(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	course, err := h.courseService.Get(r.Context(), &services.GetCourseQuery{
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching course: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	isInstructor := course.IsTaughtBy(user)
	if !isInstructor {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	module, err := h.moduleService.Get(r.Context(), &services.GetModuleQuery{
		ModuleID: moduleID,
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "module not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching module: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	pd := ContentNewPageData{
		User:          user,
		Course:        course,
		Module:        module,
		IsInstructor:  isInstructor,
		ActiveNavItem: "content",
	}

	tmpl, ok := h.templates["content_new.html"]
	if !ok {
		log.Printf("template not found: content_new.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
