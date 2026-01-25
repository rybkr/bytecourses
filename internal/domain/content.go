package domain

import (
	"time"
)

type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
)

type ContentType string

const (
	ContentTypeReading ContentType = "reading"
)

type ContentItem interface {
	Type() ContentType
}

var (
	_ ContentItem = (*Reading)(nil)
)

type BaseContentItem struct {
	ID        int64         `json:"id"`
	ModuleID  int64         `json:"module_id"`
	Title     string        `json:"title"`
	Order     int           `json:"order"`
	Status    ContentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type ReadingFormat string

const (
	ReadingFormatMarkdown ReadingFormat = "markdown"
)

type Reading struct {
	BaseContentItem
	Format  ReadingFormat `json:"format"`
	Content *string       `json:"content,omitempty"`
}

func (r *Reading) Type() ContentType {
	return ContentTypeReading
}
