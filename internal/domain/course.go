package domain

import (
	"time"
)

type CourseStatus string

const (
	CourseStatusDraft CourseStatus = "draft"
	CourseStatusLive  CourseStatus = "live"
)

type Course struct {
	ID           int64        `json:"id"`
	Title        string       `json:"title"`
	Summary      string       `json:"summary"`
	InstructorID int64        `json:"instructor_id"`
	Status       CourseStatus `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
}
