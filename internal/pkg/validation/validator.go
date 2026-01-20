package validation

import (
	"bytecourses/internal/pkg/errors"
)

type Validator struct {
	errs *errors.ValidationErrors
}

func New() *Validator {
	return &Validator{
		errs: errors.NewValidationErrors(),
	}
}

func (v *Validator) Validate(value interface{}) error {
	switch val := value.(type) {
	case interface {
		Validate(*Validator)
	}:
		val.Validate(v)

	default:
		return nil
	}

	if v.errs.HasErrors() {
		return v.errs
	}
	return nil
}

func (v *Validator) Errors() error {
	if v.errs.HasErrors() {
		return v.errs
	}
	return nil
}

func (v *Validator) Field(value interface{}, name string) *FieldValidator {
	return &FieldValidator{
		value: value,
		name:  name,
		errs:  v.errs,
	}
}
