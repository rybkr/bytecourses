package domain

import (
	"time"
)

type Enrollment struct {
	UserID     int64     `json:"user_id"`
	CourseID   int64     `json:"course_id"`
	EnrolledAt time.Time `json:"enrolled_at"`
}

func (e *Enrollment) IsForUser(userID int64) bool {
	return e.UserID == userID
}

func (e *Enrollment) IsForCourse(courseID int64) bool {
	return e.CourseID == courseID
}
