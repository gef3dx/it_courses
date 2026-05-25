package user

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository *Repository
	validate   *validator.Validate
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
		validate:   validator.New(),
	}
}

func (s *Service) List(ctx context.Context) ([]Model, error) {
	return s.repository.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Model, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) FindByEmail(ctx context.Context, email string) (*Model, error) {
	return s.repository.FindByEmail(ctx, email)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Model, error) {
	input.Email = strings.TrimSpace(input.Email)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Name = strings.TrimSpace(input.Name)
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)

	if input.Role == "" {
		input.Role = RoleStudent
	}

	if err := s.validate.Struct(input); err != nil {
		return nil, buildValidationError(err)
	}

	return s.repository.Create(ctx, input)
}

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

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.ensureNotLastAdmin(ctx, id, false)
}

func (s *Service) DeleteSelf(ctx context.Context, id int64, password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrPasswordRequired
	}

	target, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if target.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(target.PasswordHash), []byte(password)) != nil {
		return ErrInvalidPassword
	}

	return s.ensureNotLastAdmin(ctx, id, true)
}

func (s *Service) ensureNotLastAdmin(ctx context.Context, id int64, checkedUser bool) error {
	target, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if target.Role == RoleAdmin {
		adminCount, err := s.repository.CountByRole(ctx, RoleAdmin)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return ErrLastAdmin
		}
	}

	if checkedUser {
		return s.repository.Delete(ctx, id)
	}

	return s.repository.Delete(ctx, id)
}

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
	case "Role":
		return ValidationError{Message: "role must be one of student, teacher, admin"}
	default:
		return ValidationError{Message: "invalid request data"}
	}
}
