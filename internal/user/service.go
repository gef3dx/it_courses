package user

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Service содержит бизнес-логику user-модуля и валидацию входных данных.
type Service struct {
	repository *Repository
	validate   *validator.Validate
}

// NewService создаёт сервис пользователей поверх репозитория.
func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
		validate:   validator.New(),
	}
}

// List делегирует получение пользователей в слой репозитория.
func (s *Service) List(ctx context.Context) ([]Model, error) {
	return s.repository.List(ctx)
}

// GetByID возвращает пользователя по ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*Model, error) {
	return s.repository.FindByID(ctx, id)
}

// FindByEmail возвращает пользователя по email.
func (s *Service) FindByEmail(ctx context.Context, email string) (*Model, error) {
	return s.repository.FindByEmail(ctx, email)
}

// Create валидирует входные данные и создаёт пользователя через репозиторий.
func (s *Service) Create(ctx context.Context, input CreateInput) (*Model, error) {
	// Нормализуем строковые поля до запуска декларативной валидации.
	input.Email = strings.TrimSpace(input.Email)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Name = strings.TrimSpace(input.Name)
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)

	// Проверяем вход по validate-тегам в CreateInput.
	if err := s.validate.Struct(input); err != nil {
		return nil, buildValidationError(err)
	}

	return s.repository.Create(ctx, input)
}

// Update валидирует входные данные и обновляет пользователя через репозиторий.
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*Model, error) {
	if input.Email != nil {
		trimmed := strings.TrimSpace(*input.Email)
		input.Email = &trimmed
	}
	if input.Phone != nil {
		trimmed := strings.TrimSpace(*input.Phone)
		input.Phone = &trimmed
	}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}
	if input.FirstName != nil {
		trimmed := strings.TrimSpace(*input.FirstName)
		input.FirstName = &trimmed
	}
	if input.LastName != nil {
		trimmed := strings.TrimSpace(*input.LastName)
		input.LastName = &trimmed
	}

	if err := s.validate.Struct(input); err != nil {
		return nil, buildValidationError(err)
	}

	return s.repository.Update(ctx, id, input)
}

// Delete удаляет пользователя по ID через репозиторий.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}

// buildValidationError преобразует ошибки validator в понятное сообщение для API.
func buildValidationError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return ValidationError{Message: "invalid request data"}
	}

	fieldErr := validationErrors[0]

	switch fieldErr.Field() {
	case "Email":
		if fieldErr.Tag() == "email" {
			return ValidationError{Message: "email has invalid format"}
		}
		return ValidationError{Message: "email is required"}
	case "Phone":
		if fieldErr.Tag() == "numeric" {
			return ValidationError{Message: "phone must contain only digits"}
		}
		if fieldErr.Tag() == "min" || fieldErr.Tag() == "max" {
			return ValidationError{Message: "phone length must be between 10 and 20 digits"}
		}
		return ValidationError{Message: "phone is required"}
	case "Name":
		if fieldErr.Tag() == "min" || fieldErr.Tag() == "max" {
			return ValidationError{Message: "name length must be between 2 and 100 characters"}
		}
		return ValidationError{Message: "name is required"}
	case "FirstName":
		if fieldErr.Tag() == "min" || fieldErr.Tag() == "max" {
			return ValidationError{Message: "first_name length must be between 2 and 100 characters"}
		}
		return ValidationError{Message: "first_name is required"}
	case "LastName":
		if fieldErr.Tag() == "min" || fieldErr.Tag() == "max" {
			return ValidationError{Message: "last_name length must be between 2 and 100 characters"}
		}
		return ValidationError{Message: "last_name is required"}
	default:
		return ValidationError{Message: "invalid request data"}
	}
}
