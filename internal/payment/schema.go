package payment

type CreateInput struct {
	PaymentMethod string `json:"payment_method" validate:"required,min=2,max=50"`
}

type UpdateStatusInput struct {
	Status        string `json:"status" validate:"required,oneof=completed failed refunded"`
	TransactionID string `json:"transaction_id"`
}
