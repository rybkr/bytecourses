package proposal

import (
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/pkg/validation"
)

type CreateCommand struct {
	Title                string
	Summary              string
	Qualifications       string
	TargetAudience       string
	LearningObjectives   string
	Outline              string
	AssumedPrerequisites string
	AuthorID             int64
}

func (c *CreateCommand) Validate(errs *errors.ValidationErrors) {
	if err := validation.EntityID(c.AuthorID, "author_id"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}

	if err := validation.Required(c.Title, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MinLength(c.Title, 5, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.Title, 200, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
}

type UpdateCommand struct {
}

type SubmitCommand struct {
}
