package domain

import (
	"time"
)

type Module struct {
	ID        int64     `json:"id"`
	CourseID  int64     `json:"course_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
