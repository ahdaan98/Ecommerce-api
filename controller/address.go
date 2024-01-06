package controller

import (
	"ecommerce-mobilestore/initializer"
	"ecommerce-mobilestore/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProceedToCheckout(c *gin.Context) {
	// Retrieve user information from the context
	var user models.User
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user details",
		})
		return
	}
	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user details",
		})
		return
	}

	// Get address details from the request body
	var address models.Address
	if err := c.Bind(&address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid address data",
		})
		return
	}

	// Save the address for the user
	address.UserID = user.ID
	if err := initializer.DB.Create(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save address",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Address added for checkout",
		"user":    user, // Return updated user information (including the address)
	})
}

func EditAddress(c *gin.Context) {
	addressID := c.Param("id")
	var user models.User
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user details",
		})
		return
	}
	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user details",
		})
		return
	}

	// Checking address exists
	var address models.Address
	if err := initializer.DB.Where("user_id = ? AND id = ?", user.ID, addressID).First(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid address",
		})
		return
	}

	var updatedAddress models.Address
	if err := c.BindJSON(&updatedAddress); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON payload",
		})
		return
	}

	// Update the existing address with the new information
	address.Street = updatedAddress.Street
	address.City = updatedAddress.City
	address.State = updatedAddress.State
	address.Country = updatedAddress.Country
	address.PostalCode = updatedAddress.PostalCode
	address.PhoneNumber = updatedAddress.PhoneNumber

	ResponseAddress := map[string]interface{}{
		"Address id":  addressID,
		"Street":      address.Street,
		"City":        address.City,
		"State":       address.State,
		"Country":     address.Country,
		"Postal code": address.PostalCode,
		"Phone no":    address.PhoneNumber,
	}

	// Save the updated address to the database
	if err := initializer.DB.Save(&address).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update address",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Address updated successfully",
		"address": ResponseAddress,
	})
}

func PlaceOrderCOD(c *gin.Context) {
	var user models.User
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user details",
		})
		return
	}
	user, ok := userInterface.(models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user details",
		})
		return
	}

	var checkoutDetails struct {
		AddressID uint `json:"address_id"`
	}
	if err := c.BindJSON(&checkoutDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid checkout details",
		})
		return
	}

	var address models.Address
	if err := initializer.DB.First(&address, checkoutDetails.AddressID).Error; err != nil || address.UserID != user.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid address selection",
		})
		return
	}

	var cartItems []models.Cart
	if err := initializer.DB.Where("user_id = ?", user.ID).Find(&cartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve cart items",
		})
		return
	}

	if len(cartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot place an order with an empty cart",
		})
		return
	}

	// Fetch product details for each cart item
	var orderedProducts []models.ProductOrder
	totalAmount := 0.0

	for _, cartItem := range cartItems {
		var product models.Product
		if err := initializer.DB.First(&product, cartItem.ProductID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve product details",
			})
			return
		}

		totalAmount += float64(cartItem.Quantity) * product.Price

		// Check if the same product is already in the user's existing orders
		existingProductOrder := models.ProductOrder{}
		if err := initializer.DB.Preload("Order").Where("orders.user_id = ? AND product_id = ?", user.ID, product.ID).First(&existingProductOrder).Error; err == nil {
			// If the product is already in the order, update the quantity
			existingProductOrder.Quantity += int(cartItem.Quantity)
			existingProductOrder.TotalPrice = float64(existingProductOrder.Quantity) * product.Price

			// Update the existing product order in the database
			if err := initializer.DB.Save(&existingProductOrder).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update product quantity in order",
				})
				return
			}

			continue
		}

		// If the product is not in the existing orders, add it to the orderedProducts slice
		orderedProduct := models.ProductOrder{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    int(cartItem.Quantity),
			TotalPrice:  float64(cartItem.Quantity) * product.Price,
		}

		orderedProducts = append(orderedProducts, orderedProduct)
	}

	// Create a new order only if there are products not already present in the user's existing orders
	if len(orderedProducts) > 0 {
		order := models.Order{
			UserID:        user.ID,
			AddressID:     checkoutDetails.AddressID,
			TotalAmount:   totalAmount,
			PaymentMethod: "cash_on_delivery",
			Products:      orderedProducts, // Add ordered products to the order details
			Status:        "pending",       // You might want to set the initial status
		}
		if err := initializer.DB.Create(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to place order",
			})
			return
		}
	}

	// Clear the user's cart after placing the order
	if err := initializer.DB.Where("user_id = ?", user.ID).Delete(&models.Cart{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear cart",
		})
		return
	}

	fullAddress := fmt.Sprintf("%s, %s, %s, %s, %s",
		address.Street, address.City, address.State, address.Country, address.PostalCode)

	c.JSON(http.StatusOK, gin.H{
		"message": "Order placed successfully",
		"order_details": gin.H{
			"user_id":        user.ID,
			"address":        fullAddress,
			"products":       orderedProducts,
			"total_amount":   totalAmount,
			"payment_method": "cash on delivery",
			"order_status":   "pending",
		},
	})
}
