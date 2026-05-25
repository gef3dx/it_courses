package course

import "time"

type CreateInput struct {
	Title       string  `json:"title" validate:"required,min=2,max=255"`
	Slug        string  `json:"slug" validate:"omitempty,min=2,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"gte=0"`
	IsPublished bool    `json:"is_published"`
}

type UpdateInput struct {
	Title       *string  `json:"title" validate:"omitempty,min=2,max=255"`
	Slug        *string  `json:"slug" validate:"omitempty,min=2,max=255"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" validate:"omitempty,gte=0"`
	IsPublished *bool    `json:"is_published"`
}

type GrantAccessInput struct {
	UserID    int64      `json:"user_id" validate:"required,gt=0"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type ReorderResponse struct {
	Data []Model `json:"data"`
}
