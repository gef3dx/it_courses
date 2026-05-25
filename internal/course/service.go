package course

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/gef3dx/it_courses/internal/slug"
	"github.com/gef3dx/it_courses/internal/user"
)

type Service struct {
	repository *Repository
	validate   *validator.Validate
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository, validate: validator.New()}
}

func (s *Service) List(ctx context.Context, role string, userID int64) ([]Model, error) {
	if role == user.RoleStudent {
		return s.repository.ListAccessible(ctx, userID)
	}
	return s.repository.List(ctx)
}

func (s *Service) MyCourses(ctx context.Context, userID int64) ([]Model, error) {
	return s.repository.ListAccessible(ctx, userID)
}

func (s *Service) GetBySlug(ctx context.Context, slugValue, role string, userID int64) (*Model, error) {
	model, err := s.repository.FindBySlug(ctx, slugValue)
	if err != nil {
		return nil, err
	}

	if role == user.RoleStudent {
		hasAccess, err := s.repository.HasAccess(ctx, model.ID, userID)
		if err != nil {
			return nil, err
		}
		if !hasAccess {
			return nil, ErrCourseAccessDenied
		}
	}

	return model, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput, authorID int64) (*Model, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Description = strings.TrimSpace(input.Description)

	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: "invalid request data"}
	}

	slugValue := input.Slug
	if slugValue == "" {
		slugValue = slug.Make(input.Title)
	} else {
		slugValue = slug.Make(slugValue)
	}

	return s.repository.Create(ctx, input, authorID, slugValue)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*Model, error) {
	if input.Title != nil {
		value := strings.TrimSpace(*input.Title)
		input.Title = &value
	}
	if input.Slug != nil {
		value := strings.TrimSpace(*input.Slug)
		input.Slug = &value
	}
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		input.Description = &value
	}

	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: "invalid request data"}
	}

	slugValue := ""
	if input.Slug != nil {
		slugValue = slug.Make(*input.Slug)
	}

	return s.repository.Update(ctx, id, input, slugValue)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}

func (s *Service) GrantAccess(ctx context.Context, courseID int64, input GrantAccessInput, adminID int64) (*Access, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: "invalid request data"}
	}

	if _, err := s.repository.FindByID(ctx, courseID); err != nil {
		return nil, err
	}

	return s.repository.GrantAccess(ctx, courseID, input.UserID, &adminID, input.ExpiresAt)
}

func (s *Service) RevokeAccess(ctx context.Context, courseID, userID int64) error {
	return s.repository.RevokeAccess(ctx, courseID, userID)
}

func (s *Service) ListAccesses(ctx context.Context, courseID int64) ([]Access, error) {
	if _, err := s.repository.FindByID(ctx, courseID); err != nil {
		return nil, err
	}
	return s.repository.ListAccesses(ctx, courseID)
}
