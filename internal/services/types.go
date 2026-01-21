package services

import (
    "bytecourses/internal/pkg/validation"
)

type Command interface {
    Validate(v *validation.Validator)
}

type Message interface {
    Command
}
