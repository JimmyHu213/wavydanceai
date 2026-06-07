package common

import (
	"unicode"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

// IsPasswordComplexEnough enforces the user-facing password complexity rule:
// length 8–24, must contain at least one letter AND one digit. Length is
// also enforced by the `validate:"min=8,max=24"` struct tag on User.Password;
// we re-check here so callers don't depend on tag ordering or on Validate.Struct
// having already run.
func IsPasswordComplexEnough(p string) bool {
	if len(p) < 8 || len(p) > 24 {
		return false
	}
	var hasLetter, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return true
		}
	}
	return false
}
