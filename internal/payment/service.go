package payment

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/gef3dx/it_courses/internal/course"
	"github.com/gef3dx/it_courses/internal/user"
)

type Service struct {
	repository *Repository
	courseRepo *course.Repository
	validate   *validator.Validate
}

func NewService(repository *Repository, courseRepo *course.Repository) *Service {
	return &Service{repository: repository, courseRepo: courseRepo, validate: validator.New()}
}

func (s *Service) List(ctx context.Context, role string, userID int64) ([]Model, error) {
	if role == user.RoleAdmin {
		return s.repository.List(ctx)
	}
	return s.repository.ListByUser(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id int64, role string, userID int64) (*Model, error) {
	item, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role != user.RoleAdmin && item.UserID != userID {
		return nil, ErrPaymentAccessDenied
	}
	return item, nil
}

func (s *Service) Create(ctx context.Context, courseID, userID int64, input CreateInput) (*Model, error) {
	input.PaymentMethod = strings.TrimSpace(input.PaymentMethod)
	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: "invalid request data"}
	}

	courseModel, err := s.courseRepo.FindByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	hasPending, err := s.repository.HasPending(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, ErrPaymentConflict
	}

	return s.repository.Create(ctx, userID, courseModel, input)
}

func (s *Service) UpdateStatus(ctx context.Context, paymentID int64, input UpdateStatusInput) (*Model, error) {
	input.Status = strings.TrimSpace(input.Status)
	input.TransactionID = strings.TrimSpace(input.TransactionID)
	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: "invalid request data"}
	}

	current, err := s.repository.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if !isTransitionAllowed(current.Status, input.Status) {
		return nil, ErrPaymentStateConflict
	}

	if input.Status == StatusCompleted {
		return s.repository.CompleteWithAccess(ctx, paymentID, input.TransactionID)
	}

	return s.repository.UpdateStatus(ctx, paymentID, input.Status, input.TransactionID)
}

func isTransitionAllowed(current, next string) bool {
	switch current {
	case StatusPending:
		return next == StatusCompleted || next == StatusFailed
	case StatusCompleted:
		return next == StatusRefunded
	default:
		return false
	}
}
