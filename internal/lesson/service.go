package lesson

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

func NewService(repository *Repository) *Service { return &Service{repository: repository, validate: validator.New()} }

func (s *Service) ListByCourse(ctx context.Context, courseID int64) ([]Model, error) { return s.repository.ListByCourse(ctx, courseID) }
func (s *Service) GetByID(ctx context.Context, courseID, id int64) (*Model, error) { return s.repository.FindByID(ctx, courseID, id) }
func (s *Service) Create(ctx context.Context, courseID int64, input CreateInput, authorID int64) (*Model, error) {
	input.Title = strings.TrimSpace(input.Title); input.Slug = strings.TrimSpace(input.Slug)
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message:"invalid request data"} }
	value := input.Slug
	if value == "" { value = slug.Make(input.Title) } else { value = slug.Make(value) }
	return s.repository.Create(ctx, courseID, input, authorID, value)
}
func (s *Service) Update(ctx context.Context, courseID, id int64, input UpdateInput) (*Model, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message:"invalid request data"} }
	value := ""
	if input.Slug != nil { value = slug.Make(*input.Slug) }
	return s.repository.Update(ctx, courseID, id, input, value)
}
func (s *Service) Delete(ctx context.Context, courseID, id int64) error { return s.repository.Delete(ctx, courseID, id) }
func (s *Service) Reorder(ctx context.Context, courseID int64, items []ReorderItem) error {
	if err := s.validate.Var(items, "required,min=1,dive"); err != nil { return user.ValidationError{Message:"invalid request data"} }
	return s.repository.Reorder(ctx, courseID, items)
}
func (s *Service) ListPublicQuestions(ctx context.Context, lessonID int64) ([]PublicQuestion, error) {
	items, err := s.repository.ListQuestions(ctx, lessonID)
	if err != nil { return nil, err }
	out := make([]PublicQuestion,0,len(items))
	for _, q := range items {
		pq := PublicQuestion{ID:int(q.ID), Text:q.Text, SortOrder:q.SortOrder}
		for _, option := range q.Options { pq.Options = append(pq.Options, PublicAnswerOption{ID:option.ID, Text:option.Text}) }
		out = append(out, pq)
	}
	return out, nil
}
func (s *Service) CreateQuestion(ctx context.Context, lessonID int64, input QuestionInput, authorID int64) (*Question, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message:"invalid request data"} }
	if countCorrect(input.Options) != 1 { return nil, user.ValidationError{Message:"question must contain exactly one correct option"} }
	return s.repository.CreateQuestion(ctx, lessonID, input, authorID)
}
func (s *Service) UpdateQuestion(ctx context.Context, lessonID, id int64, input QuestionInput) (*Question, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message:"invalid request data"} }
	if countCorrect(input.Options) != 1 { return nil, user.ValidationError{Message:"question must contain exactly one correct option"} }
	return s.repository.UpdateQuestion(ctx, lessonID, id, input)
}
func (s *Service) DeleteQuestion(ctx context.Context, lessonID, id int64) error { return s.repository.DeleteQuestion(ctx, lessonID, id) }
func (s *Service) Submit(ctx context.Context, lessonID int64, input SubmitInput) (*SubmitResponse, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message:"invalid request data"} }
	return s.repository.Submit(ctx, lessonID, input.Answers)
}
func (s *Service) LinkTest(ctx context.Context, lessonID int64, input TestLinkInput) (*TestLink, error) {
	if err := s.validate.Struct(input); err != nil { return nil, user.ValidationError{Message:"invalid request data"} }
	return s.repository.LinkTest(ctx, lessonID, input)
}
func (s *Service) UnlinkTest(ctx context.Context, lessonID, testID int64) error { return s.repository.UnlinkTest(ctx, lessonID, testID) }

func countCorrect(items []AnswerOptionInput) int {
	count := 0
	for _, item := range items { if item.IsCorrect { count++ } }
	return count
}
