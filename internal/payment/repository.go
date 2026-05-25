package payment

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/gef3dx/it_courses/internal/course"
)

type Repository struct {
	db         *gorm.DB
	courseRepo *course.Repository
}

func NewRepository(db *gorm.DB, courseRepo *course.Repository) *Repository {
	return &Repository{db: db, courseRepo: courseRepo}
}

func (r *Repository) List(ctx context.Context) ([]Model, error) {
	var items []Model
	err := r.db.WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) ListByUser(ctx context.Context, userID int64) ([]Model, error) {
	var items []Model
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Model, error) {
	var item Model
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) HasPending(ctx context.Context, userID, courseID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).
		Where("user_id = ? AND course_id = ? AND status = ?", userID, courseID, StatusPending).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) Create(ctx context.Context, userID int64, courseModel *course.Model, input CreateInput) (*Model, error) {
	item := &Model{
		UserID:        userID,
		CourseID:      courseModel.ID,
		Amount:        courseModel.Price,
		Currency:      "RUB",
		Status:        StatusPending,
		PaymentMethod: input.PaymentMethod,
	}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, paymentID int64, status, transactionID string) (*Model, error) {
	item, err := r.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"status":         status,
		"transaction_id": transactionID,
	}
	if status == StatusCompleted {
		now := time.Now().UTC()
		updates["paid_at"] = &now
	}

	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", paymentID).Updates(updates).Error; err != nil {
		return nil, err
	}

	return r.FindByID(ctx, item.ID)
}

func (r *Repository) CompleteWithAccess(ctx context.Context, paymentID int64, transactionID string) (*Model, error) {
	var result *Model

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item Model
		if err := tx.Where("id = ?", paymentID).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}

		now := time.Now().UTC()
		if err := tx.Model(&Model{}).Where("id = ?", paymentID).Updates(map[string]interface{}{
			"status":         StatusCompleted,
			"transaction_id": transactionID,
			"paid_at":        &now,
		}).Error; err != nil {
			return err
		}

		access := &course.Access{
			CourseID: item.CourseID,
			UserID:   item.UserID,
		}
		if err := tx.Where("course_id = ? AND user_id = ?", item.CourseID, item.UserID).FirstOrCreate(access).Error; err != nil {
			return err
		}

		var updated Model
		if err := tx.Where("id = ?", paymentID).First(&updated).Error; err != nil {
			return err
		}
		result = &updated
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
