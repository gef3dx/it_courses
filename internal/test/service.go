package test

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/gef3dx/it_courses/internal/user"
)

type Service struct {
	repository *Repository
	validate   *validator.Validate
}

func NewService(repository *Repository) *Service { return &Service{repository: repository, validate: validator.New()} }

func (s *Service) List(ctx context.Context) ([]Model, error) { return s.repository.List(ctx) }
func (s *Service) GetByID(ctx context.Context, id int64) (*Model, error) { return s.repository.FindByID(ctx, id) }
func (s *Service) Create(ctx context.Context, input CreateInput, authorID int64) (*Model, error) {
	input.Title = strings.TrimSpace(input.Title)
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message: "invalid request data"} }
	return s.repository.Create(ctx, input, authorID)
}
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*Model, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message: "invalid request data"} }
	return s.repository.Update(ctx, id, input)
}
func (s *Service) Delete(ctx context.Context, id int64) error { return s.repository.Delete(ctx, id) }
func (s *Service) ListPublicQuestions(ctx context.Context, testID int64) ([]PublicQuestion, error) {
	items, err := s.repository.ListQuestionsByTest(ctx, testID)
	if err != nil { return nil, err }
	return toPublicQuestions(items, false), nil
}
func (s *Service) CreateQuestion(ctx context.Context, testID int64, input QuestionInput, authorID int64) (*Question, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message: "invalid request data"} }
	if countCorrect(input.Options) != 1 { return nil, ErrInvalidQuestionData }
	return s.repository.CreateQuestion(ctx, testID, input, authorID)
}
func (s *Service) UpdateQuestion(ctx context.Context, id int64, input QuestionInput) (*Question, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message: "invalid request data"} }
	if countCorrect(input.Options) != 1 { return nil, ErrInvalidQuestionData }
	return s.repository.UpdateQuestion(ctx, id, input)
}
func (s *Service) DeleteQuestion(ctx context.Context, id int64) error { return s.repository.DeleteQuestion(ctx, id) }
func (s *Service) Submit(ctx context.Context, testID, userID int64, input SubmitInput) (*Result, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message: "invalid request data"} }
	return s.repository.Submit(ctx, testID, userID, input.Answers)
}
func (s *Service) ListResults(ctx context.Context, role string, userID int64) ([]Result, error) {
	if role == user.RoleAdmin || role == user.RoleTeacher {
		return s.repository.ListResultsByUser(ctx, userID)
	}
	return s.repository.ListResultsByUser(ctx, userID)
}
func (s *Service) GetResultByID(ctx context.Context, id int64, role string, userID int64) (*Result, error) {
	item, err := s.repository.FindResultByID(ctx, id)
	if err != nil { return nil, err }
	if role == user.RoleAdmin || role == user.RoleTeacher || item.UserID == userID { return item, nil }
	return nil, ErrResultAccessDenied
}
func (s *Service) ListResultsByTest(ctx context.Context, testID int64) ([]Result, error) {
	return s.repository.ListResultsByTest(ctx, testID)
}

func countCorrect(options []AnswerOptionInput) int {
	count := 0
	for _, item := range options { if item.IsCorrect { count++ } }
	return count
}

func toPublicQuestions(items []Question, includeSolution bool) []PublicQuestion {
	out := make([]PublicQuestion, 0, len(items))
	for _, item := range items {
		public := PublicQuestion{ID: item.ID, Text: item.Text, SortOrder: item.SortOrder}
		if includeSolution { public.Solution = item.Solution }
		for _, option := range item.Options {
			public.Options = append(public.Options, PublicAnswerOption{ID: option.ID, Text: option.Text})
		}
		out = append(out, public)
	}
	return out
}

func IsNotFound(err error) bool { return errors.Is(err, ErrTestNotFound) || errors.Is(err, ErrQuestionNotFound) || errors.Is(err, ErrResultNotFound) }
