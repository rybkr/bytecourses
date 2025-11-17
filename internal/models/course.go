package models

import (
	"time"
)

type CourseStatus string

const (
	StatusPending  CourseStatus = "pending"
	StatusApproved CourseStatus = "approved"
	StatusRejected CourseStatus = "rejected"
)

type Course struct {
	ID              int          `json:"id"`
	InstructorID    int          `json:"instructor_id"`
	Title           string       `json:"title"`
	Description     string       `json:"description"`
	Status          CourseStatus `json:"status"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}
