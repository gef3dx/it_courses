package user

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Repository инкапсулирует прямую работу с таблицей users через GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository создаёт репозиторий пользователей поверх общего подключения к БД.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// List возвращает всех пользователей из БД в порядке возрастания id.
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

// Create создаёт новую запись пользователя в таблице users.
func (r *Repository) Create(ctx context.Context, input CreateInput) (*Model, error) {
	model := &Model{
		Email:     strings.TrimSpace(input.Email),
		Phone:     strings.TrimSpace(input.Phone),
		Name:      strings.TrimSpace(input.Name),
		FirstName: strings.TrimSpace(input.FirstName),
		LastName:  strings.TrimSpace(input.LastName),
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserConflict
		}

		return nil, err
	}

	return model, nil
}

func (r *Repository) Update(ctx context.Context, id int64, input UpdateInput) (*Model, error) {
	model := &Model{}

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
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

	if err := r.db.WithContext(ctx).Model(model).Updates(updates).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserConflict
		}
		return nil, err
	}

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(model).Error; err != nil {
		return nil, err
	}

	return model, nil
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

	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
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

	err := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &model, nil
}
