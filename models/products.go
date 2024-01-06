package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name       string  `json:"productname"`
	Price      float64 `json:"price"`
	Quantity   int     `json:"quantity"`
	CategoryID uint    `json:"category_id"`
	Available  bool
}

type Productimg struct {
	gorm.Model
	ProductID uint
	Image     string
}