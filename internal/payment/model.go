package payment

import "time"

const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusRefunded  = "refunded"
)

type Model struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id" gorm:"column:user_id"`
	CourseID      int64      `json:"course_id" gorm:"column:course_id"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method" gorm:"column:payment_method"`
	TransactionID string     `json:"transaction_id" gorm:"column:transaction_id"`
	PaidAt        *time.Time `json:"paid_at" gorm:"column:paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Model) TableName() string {
	return "payments"
}
