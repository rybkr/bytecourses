package services

import (
    "bytecourses/internal/pkg/validation"
)

type Message interface {
    Validate(v *validation.Validator)
}
