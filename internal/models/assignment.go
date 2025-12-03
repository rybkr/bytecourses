package models

import (
	"time"
)

type Assignment struct {
	ID          int       `json:"id"`
	CourseID    int       `json:"course_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Points      int       `json:"points"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Submission struct {
	ID          int       `json:"id"`
	AssignmentID int      `json:"assignment_id"`
	StudentID    int      `json:"student_id"`
	Content      string   `json:"content"`
	SubmittedAt  time.Time `json:"submitted_at"`
	Status       string   `json:"status"`
}

type Grade struct {
	ID           int       `json:"id"`
	SubmissionID int       `json:"submission_id"`
	InstructorID int       `json:"instructor_id"`
	Score        *int      `json:"score,omitempty"`
	Feedback     string    `json:"feedback"`
	GradedAt     time.Time `json:"graded_at"`
}

