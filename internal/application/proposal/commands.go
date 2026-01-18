package proposal

import (
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/pkg/validation"
)

type CreateCommand struct {
	AuthorID             int64
	Title                string
	Summary              string
	Qualifications       string
	TargetAudience       string
	LearningObjectives   string
	Outline              string
	AssumedPrerequisites string
}

func (c *CreateCommand) Validate(errs *errors.ValidationErrors) {
	if err := validation.EntityID(c.AuthorID, "author_id"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}

	if err := validation.Required(c.Title, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.Required(c.Summary, "summary"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MinLength(c.Title, 4, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}

	if err := validation.MaxLength(c.Title, 128, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.Summary, 2048, "summary"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.Qualifications, 2048, "qualifications"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
    if err := validation.MaxLength(c.TargetAudience, 2048, "target_audience"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.LearningObjectives, 2048, "learning_objectives"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.Outline, 2048, "outline"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.AssumedPrerequisites, 2048, "assumed_prerequisites"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
}

type UpdateCommand struct {
	ProposalID           int64
	UserID               int64
	Title                string
	Summary              string
	Qualifications       string
	TargetAudience       string
	LearningObjectives   string
	Outline              string
	AssumedPrerequisites string
}

func (c *UpdateCommand) Validate(errs *validation.ValidationErrors) {
    if err := validation.EntityID(c.ProposalID, "proposal_id"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
    if err := validation.UserID(c.UserID, "user_id"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}

	if err := validation.Required(c.Title, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.Required(c.Summary, "summary"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MinLength(c.Title, 4, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}

	if err := validation.MaxLength(c.Title, 128, "title"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.Summary, 2048, "summary"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.Qualifications, 2048, "qualifications"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
    if err := validation.MaxLength(c.TargetAudience, 2048, "target_audience"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.LearningObjectives, 2048, "learning_objectives"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.Outline, 2048, "outline"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
	if err := validation.MaxLength(c.AssumedPrerequisites, 2048, "assumed_prerequisites"); err != nil {
		errs.Errors = append(errs.Errors, *err)
	}
}

type SubmitCommand struct {
}
