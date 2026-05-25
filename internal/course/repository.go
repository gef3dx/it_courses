package course

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
	var models []Model
	err := r.db.WithContext(ctx).Order("id ASC").Find(&models).Error
	return models, err
}

func (r *Repository) ListAccessible(ctx context.Context, userID int64) ([]Model, error) {
	var models []Model
	err := r.db.WithContext(ctx).
		Table("courses").
		Joins("JOIN course_accesses ON course_accesses.course_id = courses.id").
		Where("course_accesses.user_id = ?", userID).
		Where("course_accesses.expires_at IS NULL OR course_accesses.expires_at > ?", time.Now().UTC()).
		Order("courses.id ASC").
		Find(&models).Error
	return models, err
}

func (r *Repository) Create(ctx context.Context, input CreateInput, authorID int64, slug string) (*Model, error) {
	model := &Model{
		Title:       strings.TrimSpace(input.Title),
		Slug:        slug,
		Description: strings.TrimSpace(input.Description),
		Price:       input.Price,
		AuthorID:    authorID,
		IsPublished: input.IsPublished,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCourseConflict
		}
		return nil, err
	}

	return model, nil
}

func (r *Repository) Update(ctx context.Context, id int64, input UpdateInput, slug string) (*Model, error) {
	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}
	if input.IsPublished != nil {
		updates["is_published"] = *input.IsPublished
	}
	if input.Slug != nil {
		updates["slug"] = slug
	}

	result := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return nil, ErrCourseConflict
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrCourseNotFound
	}

	return r.FindByID(ctx, id)
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Model, error) {
	var model Model
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCourseNotFound
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
			return nil, ErrCourseNotFound
		}
		return nil, err
	}
	return &model, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Model{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCourseNotFound
	}
	return nil
}

func (r *Repository) GrantAccess(ctx context.Context, courseID, userID int64, grantedBy *int64, expiresAt *time.Time) (*Access, error) {
	model := &Access{
		CourseID:  courseID,
		UserID:    userID,
		GrantedBy: grantedBy,
		ExpiresAt: expiresAt,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCourseAccessExists
		}
		return nil, err
	}
	return model, nil
}

func (r *Repository) RevokeAccess(ctx context.Context, courseID, userID int64) error {
	result := r.db.WithContext(ctx).Where("course_id = ? AND user_id = ?", courseID, userID).Delete(&Access{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCourseAccessMissing
	}
	return nil
}

func (r *Repository) ListAccesses(ctx context.Context, courseID int64) ([]Access, error) {
	var items []Access
	err := r.db.WithContext(ctx).Where("course_id = ?", courseID).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) HasAccess(ctx context.Context, courseID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Access{}).
		Where("course_id = ? AND user_id = ?", courseID, userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now().UTC()).
		Count(&count).Error
	return count > 0, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
