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

func CourseFromProposal(p *Proposal) *Course {
	return &Course{
		Title:        p.Title,
		Summary:      p.Summary,
		InstructorID: p.AuthorID,
		Status:       CourseStatusDraft,
	}
}

func (c *Course) IsLive() bool {
	return c.Status == CourseStatusLive
}

func (c *Course) IsTaughtBy(u *User) bool {
	return c.InstructorID == u.ID
}

func (c *Course) IsViewableBy(u *User) bool {
	return u.IsAdmin() || c.IsLive() || c.IsTaughtBy(u)
}
