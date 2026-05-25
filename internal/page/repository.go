package page

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context) ([]Model, error) {
	var items []Model
	err := r.db.WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Model, error) {
	var item Model
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateInput, slug string) (*Model, error) {
	item := &Model{Title: strings.TrimSpace(input.Title), Slug: slug, Content: input.Content, IsPublished: input.IsPublished}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if unique(err) {
			return nil, ErrPageConflict
		}
		return nil, err
	}
	return item, nil
}

func (r *Repository) Update(ctx context.Context, id int64, input UpdateInput, slug string) (*Model, error) {
	updates := map[string]any{}
	if input.Title != nil {
		updates["title"] = strings.TrimSpace(*input.Title)
	}
	if input.Slug != nil {
		updates["slug"] = slug
	}
	if input.Content != nil {
		updates["content"] = *input.Content
	}
	if input.IsPublished != nil {
		updates["is_published"] = *input.IsPublished
	}
	result := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		if unique(result.Error) {
			return nil, ErrPageConflict
		}
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrPageNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Model, error) {
	var item Model
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPageNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Model{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPageNotFound
	}
	return nil
}

func unique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
