package validation

import (
	"bytecourses/internal/pkg/errors"
	"net/mail"
	"strings"
)

type FieldValidator struct {
	value interface{}
	name  string
	errs  *errors.ValidationErrors
}

func (fv *FieldValidator) Required() *FieldValidator {
	switch val := fv.value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			fv.errs.Add(fv.name, "required")
		}
	}
	return fv
}

func (fv *FieldValidator) MinLength(min int) *FieldValidator {
	if s, ok := fv.value.(string); ok {
		if len(strings.TrimSpace(s)) < min {
			fv.errs.Add(fv.name, "must be at least "+string(rune(min))+" characters")
		}
	}
	return fv
}

func (fv *FieldValidator) MaxLength(max int) *FieldValidator {
	if s, ok := fv.value.(string); ok {
		if len(s) > max {
			fv.errs.Add(fv.name, "must be at most "+string(rune(max))+" characters")
		}
	}
	return fv
}

func (fv *FieldValidator) IsLower() *FieldValidator {
	if s, ok := fv.value.(string); ok {
		if s != strings.ToLower(s) {
			fv.errs.Add(fv.name, "must be lowercase")
		}
	}
	return fv
}

func (fv *FieldValidator) IsTrimmed() *FieldValidator {
	if s, ok := fv.value.(string); ok {
		if s != strings.TrimSpace(s) {
			fv.errs.Add(fv.name, "must not be surrounded by whitespace")
		}
	}
	return fv
}

func (fv *FieldValidator) Email() *FieldValidator {
	if s, ok := fv.value.(string); ok {
		if _, err := mail.ParseAddress(strings.TrimSpace(s)); err != nil {
			fv.errs.Add(fv.name, "invalid email format")
		}
	}
	return fv.MaxLength(254).IsLower().IsTrimmed()
}

func (fv *FieldValidator) Password() *FieldValidator {
    return fv.MinLength(1).IsTrimmed()
}

func (fv *FieldValidator) EntityID() *FieldValidator {
	if id, ok := fv.value.(int64); ok {
		if id <= 0 {
			fv.errs.Add(fv.name, "invalid entity id")
		}
	}
	return fv
}
