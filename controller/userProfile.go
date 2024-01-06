package controller

import (
	"ecommerce-mobilestore/initializer"
	"ecommerce-mobilestore/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Profile(c *gin.Context) {
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

	var fetchedUser models.User
	// Fetch the user with the specified ID
	if err := initializer.DB.First(&fetchedUser, user.ID).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	// Display user details, addresses, orders, orderedProducts, and order amounts
	c.JSON(200, gin.H{
		"email":   fetchedUser.Email,
		"phoneno": fetchedUser.Phoneno,
	})
}

func ViewAddress(c *gin.Context) {
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

	var addresses []models.Address
	if err := initializer.DB.Where("user_id = ?", user.ID).Find(&addresses).Error; err != nil {
		c.JSON(404, gin.H{"error": "Addresses not found"})
		return
	}

	var responseAddresses []map[string]interface{}

	for _, i := range addresses {
		respAddress := map[string]interface{}{
			"address id":  i.ID,
			"Street":      i.Street,
			"City":        i.City,
			"Country":     i.Country,
			"Postal Code": i.PostalCode,
		}

		responseAddresses = append(responseAddresses, respAddress)
	}

	c.JSON(200, gin.H{
		"address details": responseAddresses,
	})
}

func ViewOrders(c *gin.Context) {
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

	// Fetch orders associated with the user and preload product details
	var orders []models.Order
	if err := initializer.DB.Preload("Products").Where("user_id = ?", user.ID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user orders",
		})
		return
	}

	var responseOrders []map[string]interface{}

	for _, order := range orders {
		var responseProducts []map[string]interface{}

		for _, product := range order.Products {
			// Retrieve associated product images
			var productImages []models.Productimg
			if err := initializer.DB.Where("product_id = ?", product.ProductID).Find(&productImages).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to retrieve product images",
				})
				return
			}

			// Extract image names
			var imageNames []string
			for _, img := range productImages {
				imageNames = append(imageNames, img.Image)
			}

			respProduct := map[string]interface{}{
				"product_id":    product.ProductID,
				"product_name":  product.ProductName,
				"quantity":      product.Quantity,
				"product_images": imageNames,
				// Add other product fields as needed
			}
			responseProducts = append(responseProducts, respProduct)
		}

		respOrder := map[string]interface{}{
			"order_id":       order.ID,
			"address_id":     order.AddressID,
			"order_status":   order.Status,
			"payment_method": order.PaymentMethod,
			"products":       responseProducts,
			"total_amount":   order.TotalAmount,
			"ordered_on":     order.CreatedAt,
		}

		responseOrders = append(responseOrders, respOrder)
	}

	// Display the list of orders made by the user
	c.JSON(http.StatusOK, gin.H{
		"order_details": responseOrders,
	})
}

func EditProfile(c *gin.Context){
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
 
	initializer.DB.First(&user, "id = ?", user.ID)
	if user.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User is not existed",
		})
		return
	}

	var updatedProfile struct {
		Email    string
		Phoneno  string
		Password string
	}

	err := c.Bind(&updatedProfile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	if len(updatedProfile.Password) < 8 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Password must contain 8 Letters",
		})
		return
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(updatedProfile.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	user.Email = updatedProfile.Email
	user.Phoneno = updatedProfile.Phoneno
	user.Password = string(hash)
	user.Blocked = false

	// Save the updated user to the database
	if err := initializer.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user profile",
		})
		return
	}

	response:=map[string]interface{}{
		"email":user.Email,
		"phone no":user.Phoneno,
		"password":"XXXXXXXXXXX",
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    response,
	})
}