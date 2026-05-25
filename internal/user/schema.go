package user

// CreateInput описывает JSON-тело запроса на создание пользователя.
type CreateInput struct {
	Email     string `json:"email" validate:"required,email" example:"ivan@example.com"`
	Phone     string `json:"phone" validate:"required,numeric,min=10,max=20" example:"9094445566"`
	Name      string `json:"name" validate:"required,min=2,max=100" example:"Ivan"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100" example:"Ivanov"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100" example:"Ivanovich"`
}

// UpdateInput описывает JSON-тело запроса на обновление пользователя.
// Все поля опциональны — обновляются только переданные.
type UpdateInput struct {
	Email     *string `json:"email" validate:"omitempty,email" example:"ivan@example.com"`
	Phone     *string `json:"phone" validate:"omitempty,numeric,min=10,max=20" example:"9094445566"`
	Name      *string `json:"name" validate:"omitempty,min=2,max=100" example:"Ivan"`
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=100" example:"Ivanov"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=100" example:"Ivanovich"`
}

// ErrorResponse используется для возврата текстовой ошибки из user-обработчиков.
type ErrorResponse struct {
	Error string `json:"error" example:"email is required"`
}
