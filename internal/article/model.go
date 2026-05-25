package article

import "time"

type Model struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Slug        string      `json:"slug"`
	Content     string      `json:"content"`
	AuthorID    int64       `json:"author_id" gorm:"column:author_id"`
	IsPublished bool        `json:"is_published" gorm:"column:is_published"`
	Media       []Media     `json:"media" gorm:"foreignKey:ArticleID"`
	Tests       []TestLink  `json:"tests" gorm:"foreignKey:ArticleID"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (Model) TableName() string { return "articles" }

type Media struct {
	ID        int64  `json:"id"`
	ArticleID int64  `json:"article_id" gorm:"column:article_id"`
	MediaType string `json:"media_type" gorm:"column:media_type"`
	URL       string `json:"url"`
	Caption   string `json:"caption"`
	SortOrder int    `json:"sort_order" gorm:"column:sort_order"`
}

func (Media) TableName() string { return "article_media" }

type TestLink struct {
	ID          int64  `json:"id"`
	ArticleID   int64  `json:"article_id" gorm:"column:article_id"`
	TestID      int64  `json:"test_id" gorm:"column:test_id"`
	Description string `json:"description"`
}

func (TestLink) TableName() string { return "article_test_links" }
