package models

import (
	"time"
)

type ApplicationStatus string

const (
	StatusDraft    ApplicationStatus = "draft"
	StatusPending  ApplicationStatus = "pending"
	StatusRejected ApplicationStatus = "rejected"
)

type Application struct {
	ID                      int              `json:"id"`
	InstructorID            int              `json:"instructor_id"`
	Title                   string           `json:"title"`
	Description             string           `json:"description"`
	LearningObjectives      string           `json:"learning_objectives"`
	Prerequisites           string           `json:"prerequisites"`
	CourseFormat            string           `json:"course_format"`
	CategoryTags            string           `json:"category_tags"`
	SkillLevel              string           `json:"skill_level"`
	CourseDuration          string           `json:"course_duration"`
	InstructorQualifications string          `json:"instructor_qualifications"`
	Status                  ApplicationStatus `json:"status"`
	RejectedAt              *time.Time       `json:"rejected_at,omitempty"`
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at"`
}

