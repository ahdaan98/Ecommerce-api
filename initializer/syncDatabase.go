package initializer

import "ecommerce-mobilestore/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Admin{})
	DB.AutoMigrate(&models.Product{})
	DB.AutoMigrate(&models.Productimg{})
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Cart{})
	DB.AutoMigrate(&models.Address{})
	DB.AutoMigrate(&models.Order{})
	DB.AutoMigrate(&models.ProductOrder{})
}