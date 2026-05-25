package lesson

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
	SortOrder   int             `json:"sort_order"`
	IsPublished bool            `json:"is_published"`
	Media       []MediaInput    `json:"media"`
	Tests       []TestLinkInput `json:"tests"`
}

type UpdateInput struct {
	Title       *string          `json:"title" validate:"omitempty,min=2,max=255"`
	Slug        *string          `json:"slug" validate:"omitempty,min=2,max=255"`
	Content     *string          `json:"content"`
	SortOrder   *int             `json:"sort_order"`
	IsPublished *bool            `json:"is_published"`
	Media       *[]MediaInput    `json:"media"`
	Tests       *[]TestLinkInput `json:"tests"`
}

type ReorderItem struct {
	ID        int64 `json:"id" validate:"required,gt=0"`
	SortOrder int   `json:"sort_order"`
}

type AnswerOptionInput struct {
	Text      string `json:"text" validate:"required,min=1,max=500"`
	IsCorrect bool   `json:"is_correct"`
}

type QuestionInput struct {
	Text      string              `json:"text" validate:"required,min=1"`
	Solution  string              `json:"solution"`
	SortOrder int                 `json:"sort_order"`
	Options   []AnswerOptionInput `json:"options" validate:"required,min=2,dive"`
}

type SubmitAnswerInput struct {
	QuestionID       int64 `json:"question_id" validate:"required,gt=0"`
	SelectedOptionID int64 `json:"selected_option_id" validate:"required,gt=0"`
}

type SubmitInput struct {
	Answers []SubmitAnswerInput `json:"answers" validate:"required,min=1,dive"`
}

type PublicQuestion struct {
	ID        int                  `json:"id"`
	Text      string               `json:"text"`
	Solution  string               `json:"solution,omitempty"`
	SortOrder int                  `json:"sort_order"`
	Options   []PublicAnswerOption `json:"options"`
}

type PublicAnswerOption struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}
