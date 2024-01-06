package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserID        uint
	AddressID     uint
	TotalAmount   float64
	PaymentMethod string
	Products      []ProductOrder
	Status        string
}

type ProductOrder struct {
	gorm.Model
	OrderID     uint    `json:"order_id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	TotalPrice  float64 `json:"total_price"`
}