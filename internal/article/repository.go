package article

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
	err := r.db.WithContext(ctx).Preload("Media").Preload("Tests").Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*Model, error) {
	var item Model
	err := r.db.WithContext(ctx).Preload("Media").Preload("Tests").Where("slug = ?", slug).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, input CreateInput, slug string, authorID int64) (*Model, error) {
	item := &Model{
		Title:       strings.TrimSpace(input.Title),
		Slug:        slug,
		Content:     input.Content,
		AuthorID:    authorID,
		IsPublished: input.IsPublished,
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			if unique(err) {
				return ErrArticleConflict
			}
			return err
		}
		for _, media := range input.Media {
			if err := tx.Create(&Media{
				ArticleID: item.ID, MediaType: media.MediaType, URL: media.URL, Caption: media.Caption, SortOrder: media.SortOrder,
			}).Error; err != nil {
				return err
			}
		}
		for _, link := range input.Tests {
			if err := tx.Create(&TestLink{
				ArticleID: item.ID, TestID: link.TestID, Description: link.Description,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, item.ID)
}

func (r *Repository) Update(ctx context.Context, id int64, input UpdateInput, slug string) (*Model, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		if len(updates) > 0 {
			result := tx.Model(&Model{}).Where("id = ?", id).Updates(updates)
			if result.Error != nil {
				if unique(result.Error) {
					return ErrArticleConflict
				}
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrArticleNotFound
			}
		} else {
			var count int64
			if err := tx.Model(&Model{}).Where("id = ?", id).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return ErrArticleNotFound
			}
		}

		if input.Media != nil {
			if err := tx.Where("article_id = ?", id).Delete(&Media{}).Error; err != nil {
				return err
			}
			for _, media := range *input.Media {
				if err := tx.Create(&Media{
					ArticleID: id, MediaType: media.MediaType, URL: media.URL, Caption: media.Caption, SortOrder: media.SortOrder,
				}).Error; err != nil {
					return err
				}
			}
		}
		if input.Tests != nil {
			if err := tx.Where("article_id = ?", id).Delete(&TestLink{}).Error; err != nil {
				return err
			}
			for _, link := range *input.Tests {
				if err := tx.Create(&TestLink{
					ArticleID: id, TestID: link.TestID, Description: link.Description,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Model, error) {
	var item Model
	err := r.db.WithContext(ctx).Preload("Media").Preload("Tests").Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound
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
		return ErrArticleNotFound
	}
	return nil
}

func unique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
