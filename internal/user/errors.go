package user

import "errors"

// Ошибки модуля user, которые безопасно возвращать между слоями приложения.
var (
	ErrUserConflict = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

// ValidationError описывает ошибку проверки пользовательского ввода.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// IsValidationError проверяет, относится ли ошибка к валидации входных данных.
func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}
