package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/gef3dx/it_courses/internal/user"
)

type Repository struct {
	db       *gorm.DB
	userRepo *user.Repository
}

func NewRepository(db *gorm.DB, userRepo *user.Repository) *Repository {
	return &Repository{db: db, userRepo: userRepo}
}

func (r *Repository) UserRepository() *user.Repository {
	return r.userRepo
}

func (r *Repository) CreatePasswordResetToken(ctx context.Context, userID int64, token string, expiresAt time.Time) (*PasswordResetToken, error) {
	model := &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return model, nil
}

func (r *Repository) FindPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	var model PasswordResetToken

	err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}

		return nil, err
	}

	return &model, nil
}

func (r *Repository) DeletePasswordResetToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&PasswordResetToken{}).Error
}
