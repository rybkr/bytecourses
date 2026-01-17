package domain

import (
	"time"
)

type ContentType string

const (
	ContentTypeLecture ContentType = "lecture"
)

type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
)

type ContentItem struct {
	ID        int64         `json:"id"`
	ModuleID  int64         `json:"module_id"`
	Title     string        `json:"title"`
	Type      ContentType   `json:"type"`
	Status    ContentStatus `json:"status"`
	Position  int           `json:"position"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type Lecture struct {
	ContentItemID int64  `json:"content_item_id"`
	Content       string `json:"content"`
}
