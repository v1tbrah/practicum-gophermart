package model

import "time"

type Order struct {
	UploadedAt time.Time `json:"uploaded_at"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual"`
	UserID     int64     `json:"-"`
}

type OrderStatus int

const (
	OrderStatusNew OrderStatus = iota
	OrderStatusProcessing
	OrderStatusInvalid
	OrderStatusProcessed
)

func (o OrderStatus) String() string {
	return [...]string{"NEW", "PROCESSING", "INVALID", "PROCESSED"}[o]
}
