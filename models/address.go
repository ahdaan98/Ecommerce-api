package models

import "gorm.io/gorm"

type Address struct {
	gorm.Model
	UserID      uint   // Foreign key to associate with a user
	Street      string // Street address
	City        string // City
	State       string // State or region
	PostalCode  string // Postal code or ZIP code
	Country     string // Country
	PhoneNumber string // Contact phone number
}