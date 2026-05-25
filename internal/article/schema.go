package article

type MediaInput struct {
	MediaType string `json:"media_type" validate:"required,oneof=image video"`
	URL       string `json:"url" validate:"required,url"`
	Caption   string `json:"caption"`
	SortOrder int    `json:"sort_order"`
}

type TestLinkInput struct {
	TestID      int64  `json:"test_id" validate:"required,gt=0"`
	Description string `json:"description"`
}

type CreateInput struct {
	Title       string          `json:"title" validate:"required,min=2,max=255"`
	Slug        string          `json:"slug" validate:"omitempty,min=2,max=255"`
	Content     string          `json:"content"`
	IsPublished bool            `json:"is_published"`
	Media       []MediaInput    `json:"media"`
	Tests       []TestLinkInput `json:"tests"`
}

type UpdateInput struct {
	Title       *string          `json:"title" validate:"omitempty,min=2,max=255"`
	Slug        *string          `json:"slug" validate:"omitempty,min=2,max=255"`
	Content     *string          `json:"content"`
	IsPublished *bool            `json:"is_published"`
	Media       *[]MediaInput    `json:"media"`
	Tests       *[]TestLinkInput `json:"tests"`
}
