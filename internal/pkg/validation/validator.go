package validation

import (
	"bytecourses/internal/pkg/errors"
	"strconv"
	"strings"
)

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(value interface{}) error {
	errs := errors.NewValidationErrors()

	switch val := value.(type) {
	case interface {
		Validate(*errors.ValidationErrors)
	}:
		val.Validate(errs)

	default:
		return nil
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

func Required(value string, field string) *errors.ValidationError {
	if strings.TrimSpace(value) == "" {
		return &errors.ValidationError{
			Field:   field,
			Message: "required",
		}
	}
	return nil
}

func MinLength(value string, min int, field string) *errors.ValidationError {
	if len(strings.TrimSpace(value)) < min {
		return &errors.ValidationError{
			Field:   field,
			Message: "must be at least " + strconv.Itoa(min) + " characters",
		}
	}
	return nil
}

func MaxLength(value string, max int, field string) *errors.ValidationError {
	if len(value) > max {
		return &errors.ValidationError{
			Field:   field,
			Message: "must be at most " + strconv.Itoa(max) + " characters",
		}
	}
	return nil
}

func Email(value string, field string) *errors.ValidationError {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, "@") {
		return &errors.ValidationError{
			Field:   field,
			Message: "invalid email format",
		}
	}
	return nil
}

func EntityID(value int64, field string) *errors.ValidationError {
	if value <= 0 {
		return &errors.ValidationError{
			Field:   field,
			Message: "must be a valid id",
		}
	}
	return nil
}
