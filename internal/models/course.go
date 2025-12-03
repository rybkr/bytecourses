package models

import (
	"time"
)

type Course struct {
	ID           int       `json:"id"`
	InstructorID int       `json:"instructor_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Content      string    `json:"content,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
