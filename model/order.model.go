package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	Id        uint64      `json:"id"`
	UserId    uuid.UUID   `json:"user_id"`
	Status    OrderStatus `json:"status"`
	Items     []Item      `json:"items"`
	CreatedAt *time.Time  `json:"created_at"`
	UpdatedAt *time.Time  `json:"updated_at"`
}

type Item struct {
	Id    uuid.UUID `json:"id"`
	Qty   uint      `json:"qty"`
	Price uint      `json:"price"`
}

type OrderStatus string

const (
	Accpeted  OrderStatus = "Accpeted"
	Delivered OrderStatus = "Delivered"
	Completed OrderStatus = "Completed"
)
