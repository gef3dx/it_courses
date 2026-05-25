package test

import "time"

type Model struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	AuthorID    int64      `json:"author_id" gorm:"column:author_id"`
	Questions   []Question `json:"questions,omitempty" gorm:"foreignKey:TestID"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Model) TableName() string { return "tests" }

type Question struct {
	ID        int64          `json:"id"`
	TestID    *int64         `json:"test_id,omitempty" gorm:"column:test_id"`
	LessonID  *int64         `json:"lesson_id,omitempty" gorm:"column:lesson_id"`
	Text      string         `json:"text"`
	Solution  string         `json:"solution"`
	AuthorID  int64          `json:"author_id" gorm:"column:author_id"`
	SortOrder int            `json:"sort_order" gorm:"column:sort_order"`
	Options   []AnswerOption `json:"options" gorm:"foreignKey:QuestionID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (Question) TableName() string { return "questions" }

type AnswerOption struct {
	ID         int64     `json:"id"`
	QuestionID int64     `json:"question_id" gorm:"column:question_id"`
	Text       string    `json:"text"`
	IsCorrect  bool      `json:"is_correct" gorm:"column:is_correct"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AnswerOption) TableName() string { return "answer_options" }

type Result struct {
	ID          int64        `json:"id"`
	TestID      int64        `json:"test_id" gorm:"column:test_id"`
	UserID      int64        `json:"user_id" gorm:"column:user_id"`
	Score       int          `json:"score"`
	MaxScore    int          `json:"max_score" gorm:"column:max_score"`
	Grade       float64      `json:"grade"`
	CompletedAt time.Time    `json:"completed_at" gorm:"column:completed_at"`
	Answers     []ResultItem `json:"answers,omitempty" gorm:"foreignKey:ResultID"`
	CreatedAt   time.Time    `json:"created_at"`
}

func (Result) TableName() string { return "test_results" }

type ResultItem struct {
	ID               int64 `json:"id"`
	ResultID         int64 `json:"result_id" gorm:"column:result_id"`
	QuestionID       int64 `json:"question_id" gorm:"column:question_id"`
	SelectedOptionID int64 `json:"selected_option_id" gorm:"column:selected_option_id"`
	IsCorrect        bool  `json:"is_correct" gorm:"column:is_correct"`
}

func (ResultItem) TableName() string { return "test_answers" }
