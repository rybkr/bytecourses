package models

import (
	"time"
)

type Enrollment struct {
	ID             int        `json:"id"`
	StudentID      int        `json:"student_id"`
	CourseID       int        `json:"course_id"`
	EnrolledAt     time.Time  `json:"enrolled_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
}
