package course

import "time"

type Model struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	AuthorID    int64     `json:"author_id" gorm:"column:author_id"`
	IsPublished bool      `json:"is_published" gorm:"column:is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Model) TableName() string {
	return "courses"
}

type Access struct {
	ID        int64      `json:"id"`
	CourseID  int64      `json:"course_id" gorm:"column:course_id"`
	UserID    int64      `json:"user_id" gorm:"column:user_id"`
	GrantedBy *int64     `json:"granted_by" gorm:"column:granted_by"`
	GrantedAt time.Time  `json:"granted_at" gorm:"column:granted_at"`
	ExpiresAt *time.Time `json:"expires_at" gorm:"column:expires_at"`
}

func (Access) TableName() string {
	return "course_accesses"
}
