// Package user содержит доменную логику для работы с пользователями.
package user

import "time"

// Model описывает запись пользователя в таблице users и формат ответа API.
type Model struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email" gorm:"unique"`
	Phone     string    `json:"phone"`
	Name      string    `json:"name"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName явно задаёт имя таблицы для модели пользователя в GORM.
func (Model) TableName() string {
	return "users"
}
