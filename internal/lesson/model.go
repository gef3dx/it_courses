package lesson

import "time"

type Model struct {
	ID          int64      `json:"id"`
	CourseID    int64      `json:"course_id" gorm:"column:course_id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	AuthorID    int64      `json:"author_id" gorm:"column:author_id"`
	SortOrder   int        `json:"sort_order" gorm:"column:sort_order"`
	IsPublished bool       `json:"is_published" gorm:"column:is_published"`
	Media       []Media    `json:"media" gorm:"foreignKey:LessonID"`
	Tests       []TestLink `json:"tests" gorm:"foreignKey:LessonID"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Model) TableName() string { return "lessons" }

type Media struct {
	ID        int64  `json:"id"`
	LessonID  int64  `json:"lesson_id" gorm:"column:lesson_id"`
	MediaType string `json:"media_type" gorm:"column:media_type"`
	URL       string `json:"url"`
	Caption   string `json:"caption"`
	SortOrder int    `json:"sort_order" gorm:"column:sort_order"`
}

func (Media) TableName() string { return "lesson_media" }

type TestLink struct {
	ID          int64  `json:"id"`
	LessonID    int64  `json:"lesson_id" gorm:"column:lesson_id"`
	TestID      int64  `json:"test_id" gorm:"column:test_id"`
	Description string `json:"description"`
}

func (TestLink) TableName() string { return "lesson_test_links" }

type Question struct {
	ID        int64          `json:"id"`
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
	ID         int64  `json:"id"`
	QuestionID int64  `json:"question_id" gorm:"column:question_id"`
	Text       string `json:"text"`
	IsCorrect  bool   `json:"is_correct" gorm:"column:is_correct"`
}

func (AnswerOption) TableName() string { return "answer_options" }

type SubmitResponse struct {
	Score     int                `json:"score"`
	MaxScore  int                `json:"max_score"`
	Grade     float64            `json:"grade"`
	Questions []PublicQuestion   `json:"questions"`
}
