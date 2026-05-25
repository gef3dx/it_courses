package test

type CreateInput struct {
	Title       string `json:"title" validate:"required,min=2,max=255"`
	Description string `json:"description"`
}

type UpdateInput struct {
	Title       *string `json:"title" validate:"omitempty,min=2,max=255"`
	Description *string `json:"description"`
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
	QuestionID      int64 `json:"question_id" validate:"required,gt=0"`
	SelectedOptionID int64 `json:"selected_option_id" validate:"required,gt=0"`
}

type SubmitInput struct {
	Answers []SubmitAnswerInput `json:"answers" validate:"required,min=1,dive"`
}

type PublicQuestion struct {
	ID        int64                `json:"id"`
	Text      string               `json:"text"`
	Solution  string               `json:"solution,omitempty"`
	SortOrder int                  `json:"sort_order"`
	Options   []PublicAnswerOption `json:"options"`
}

type PublicAnswerOption struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}
