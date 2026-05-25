package page

type CreateInput struct {
	Title       string `json:"title" validate:"required,min=2,max=255"`
	Slug        string `json:"slug" validate:"omitempty,min=2,max=255"`
	Content     string `json:"content"`
	IsPublished bool   `json:"is_published"`
}

type UpdateInput struct {
	Title       *string `json:"title" validate:"omitempty,min=2,max=255"`
	Slug        *string `json:"slug" validate:"omitempty,min=2,max=255"`
	Content     *string `json:"content"`
	IsPublished *bool   `json:"is_published"`
}
