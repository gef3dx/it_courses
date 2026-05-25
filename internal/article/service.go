package article

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

func (s *Service) List(ctx context.Context) ([]Model, error) { return s.repository.List(ctx) }
func (s *Service) GetBySlug(ctx context.Context, slugValue string) (*Model, error) {
	return s.repository.FindBySlug(ctx, slugValue)
}
func (s *Service) Create(ctx context.Context, input CreateInput, authorID int64) (*Model, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)
	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: "invalid request data"}
	}
	value := input.Slug
	if value == "" {
		value = slug.Make(input.Title)
	} else {
		value = slug.Make(value)
	}
	return s.repository.Create(ctx, input, value, authorID)
}
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*Model, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: "invalid request data"}
	}
	value := ""
	if input.Slug != nil {
		value = slug.Make(*input.Slug)
	}
	return s.repository.Update(ctx, id, input, value)
}
func (s *Service) Delete(ctx context.Context, id int64) error { return s.repository.Delete(ctx, id) }
