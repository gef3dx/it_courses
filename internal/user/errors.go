package user

import "errors"

var (
	ErrUserConflict      = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrLastAdmin         = errors.New("cannot delete the last admin")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrPasswordRequired  = errors.New("password is required")
	ErrForbiddenRoleEdit = errors.New("forbidden role update")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}
