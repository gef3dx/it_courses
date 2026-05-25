package user

import "time"

const (
	RoleStudent = "student"
	RoleTeacher = "teacher"
	RoleAdmin   = "admin"
)

type Model struct {
	ID                     int64      `json:"id"`
	Email                  string     `json:"email" gorm:"unique"`
	Phone                  string     `json:"phone"`
	Name                   string     `json:"name"`
	FirstName              string     `json:"first_name"`
	LastName               string     `json:"last_name"`
	PasswordHash           string     `json:"-" gorm:"column:password_hash"`
	Role                   string     `json:"role"`
	EmailVerifiedAt        *time.Time `json:"email_verified_at" gorm:"column:email_verified_at"`
	EmailVerificationToken string     `json:"-" gorm:"column:email_verification_token"`
	TokenVersion           int        `json:"-" gorm:"column:token_version"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func (Model) TableName() string {
	return "users"
}
