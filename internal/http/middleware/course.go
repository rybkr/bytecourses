package middleware

import (
	"bytecourses/internal/store"
	"net/http"
)

type CourseIDFunc func(r *http.Request) (int64, bool)

func RequireCourse(courses store.CourseStore, courseID CourseIDFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cid, ok := courseID(r)
			if !ok {
				http.Error(w, "missing id", http.StatusBadRequest)
				return
			}

			c, ok := courses.GetCourseByID(r.Context(), cid)
			if !ok {
				http.Error(w, "course not found", http.StatusNotFound)
				return
			}

			next.ServeHTTP(w, r.WithContext(withCourse(r.Context(), c)))
		})
	}
}
