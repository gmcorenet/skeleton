package validation

import "errors"

var (
	ErrNameRequired  = errors.New("name is required")
	ErrEmailInvalid = errors.New("email is invalid")
)

type UserValidator struct{}

func NewUserValidator() *UserValidator {
	return &UserValidator{}
}

func (v *UserValidator) ValidateCreate(data interface{}) error {
	// TODO: Implement validation
	return nil
}

func (v *UserValidator) ValidateUpdate(data interface{}) error {
	// TODO: Implement validation
	return nil
}
