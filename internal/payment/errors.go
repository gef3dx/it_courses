package payment

import "errors"

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentConflict      = errors.New("pending payment already exists")
	ErrInvalidStatus        = errors.New("invalid payment status")
	ErrPaymentAccessDenied  = errors.New("payment access denied")
	ErrPaymentStateConflict = errors.New("payment status transition is not allowed")
)
