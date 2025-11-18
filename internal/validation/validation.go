package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(field, message string) {
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}

func (v *Validator) Required(value string, field string) {
	v.Check(strings.TrimSpace(value) != "", field, "This field is required")
}

func (v *Validator) MinLength(value string, n int, field string) {
	v.Check(len(value) >= n, field, fmt.Sprintf("This field must be at least %d characters", n))
}

func (v *Validator) MaxLength(value string, n int, field string) {
	v.Check(len(value) <= n, field, fmt.Sprintf("This field must be no more than %d characters", n))
}

func (v *Validator) Email(value string, field string) {
	v.Check(emailRegex.MatchString(value), field, "Must be a valid email address")
}

func (v *Validator) Error() error {
	if v.Valid() {
		return nil
	}

	var messages []string
	for field, msg := range v.Errors {
		messages = append(messages, field+": "+msg)
	}
	return errors.New(strings.Join(messages, ", "))
}
