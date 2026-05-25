package user

import "time"

type CreateInput struct {
	Email                  string     `json:"email" validate:"required,email" example:"ivan@example.com"`
	Phone                  string     `json:"phone" validate:"required,numeric,min=10,max=20" example:"9094445566"`
	Name                   string     `json:"name" validate:"required,min=2,max=100" example:"Ivan"`
	FirstName              string     `json:"first_name" validate:"required,min=2,max=100" example:"Ivanov"`
	LastName               string     `json:"last_name" validate:"required,min=2,max=100" example:"Ivanovich"`
	PasswordHash           string     `json:"-" validate:"-"`
	Role                   string     `json:"role,omitempty" validate:"omitempty,oneof=student teacher admin"`
	EmailVerifiedAt        *time.Time `json:"email_verified_at,omitempty" validate:"-"`
	EmailVerificationToken string     `json:"-" validate:"-"`
	TokenVersion           int        `json:"-" validate:"gte=0"`
}

type UpdateInput struct {
	Email     *string `json:"email" validate:"omitempty,email" example:"ivan@example.com"`
	Phone     *string `json:"phone" validate:"omitempty,numeric,min=10,max=20" example:"9094445566"`
	Name      *string `json:"name" validate:"omitempty,min=2,max=100" example:"Ivan"`
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=100" example:"Ivanov"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=100" example:"Ivanovich"`
}

type DeleteInput struct {
	Password string `json:"password"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"email is required"`
}
