package domain

import (
	"time"
)

type ModuleStatus string

const (
	ModuleStatusDraft     ModuleStatus = "draft"
	ModuleStatusPublished ModuleStatus = "published"
)

type Module struct {
	ID          int64        `json:"id"`
	CourseID    int64        `json:"course_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Order       int          `json:"order"`
	Status      ModuleStatus `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
