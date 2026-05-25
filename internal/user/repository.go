package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Model, error) {
	var users []Model

	err := r.db.WithContext(ctx).
		Order("id ASC").
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) Create(ctx context.Context, input CreateInput) (*Model, error) {
	role := input.Role
	if role == "" {
		role = RoleStudent
	}

	model := &Model{
		Email:                  strings.TrimSpace(input.Email),
		Phone:                  strings.TrimSpace(input.Phone),
		Name:                   strings.TrimSpace(input.Name),
		FirstName:              strings.TrimSpace(input.FirstName),
		LastName:               strings.TrimSpace(input.LastName),
		PasswordHash:           input.PasswordHash,
		Role:                   role,
		EmailVerifiedAt:        input.EmailVerifiedAt,
		EmailVerificationToken: input.EmailVerificationToken,
		TokenVersion:           input.TokenVersion,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserConflict
		}

		return nil, err
	}

	return model, nil
}

func (r *Repository) Update(ctx context.Context, id int64, input UpdateInput) (*Model, error) {
	model, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.Email != nil {
		updates["email"] = strings.TrimSpace(*input.Email)
	}
	if input.Phone != nil {
		updates["phone"] = strings.TrimSpace(*input.Phone)
	}
	if input.Name != nil {
		updates["name"] = strings.TrimSpace(*input.Name)
	}
	if input.FirstName != nil {
		updates["first_name"] = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		updates["last_name"] = strings.TrimSpace(*input.LastName)
	}

	if len(updates) == 0 {
		return model, nil
	}

	if err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserConflict
		}

		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Model{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*Model, error) {
	var model Model

	err := r.db.WithContext(ctx).Where("email = ?", strings.TrimSpace(email)).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &model, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Model, error) {
	var model Model

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &model, nil
}

func (r *Repository) FindByVerificationToken(ctx context.Context, token string) (*Model, error) {
	var model Model

	err := r.db.WithContext(ctx).Where("email_verification_token = ?", token).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return &model, nil
}

func (r *Repository) MarkEmailVerified(ctx context.Context, id int64) (*Model, error) {
	now := time.Now().UTC()

	err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Updates(map[string]interface{}{
		"email_verified_at":        &now,
		"email_verification_token": "",
	}).Error
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *Repository) SetVerificationToken(ctx context.Context, id int64, token string) error {
	return r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Updates(map[string]interface{}{
		"email_verification_token": token,
		"email_verified_at":        nil,
	}).Error
}

func (r *Repository) UpdatePasswordHash(ctx context.Context, id int64, passwordHash string, incrementTokenVersion bool) error {
	updates := map[string]interface{}{
		"password_hash": passwordHash,
	}
	if incrementTokenVersion {
		updates["token_version"] = gorm.Expr("token_version + 1")
	}

	return r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) CountByRole(ctx context.Context, role string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).Where("role = ?", role).Count(&count).Error
	return count, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
