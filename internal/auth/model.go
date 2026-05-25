package auth

import "time"

type PasswordResetToken struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"column:user_id"`
	Token     string    `gorm:"unique"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

type Claims struct {
	UserID       int64  `json:"sub"`
	Role         string `json:"role"`
	TokenVersion int    `json:"token_version"`
	TokenType    string `json:"type"`
	ExpiresAt    int64  `json:"exp"`
}
