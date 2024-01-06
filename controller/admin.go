package controller

import (
	"ecommerce-mobilestore/initializer"
	"ecommerce-mobilestore/models"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var adminLoggedStatus = false

func AdminLogged(c *gin.Context) {
	adminLoggedStatus = true
}

func Adminlogin(c *gin.Context) {
	var body struct {
		Username string
		Password string
	}

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	var admin models.Admin
	initializer.DB.First(&admin, "username = ?", body.Username)

	if admin.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Admin name",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": admin.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "error signing admin token",
		})
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("admintoken", tokenString, 3600, "", "", false, true)
	adminLoggedStatus = true
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged in",
	})
}

func Adminlogout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("admintoken", "", -1, "", "", false, true)
	adminLoggedStatus = false
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully Log out",
	})
}

func ListofUsers(c *gin.Context) {
	ok := adminLoggedStatus
	if ok {
		var users []models.User
		initializer.DB.Find(&users)
		c.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized access",
		})
	}
}

func BlockUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	initializer.DB.First(&user, userID)
	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"ERROR": fmt.Sprintf("User with ID %s not found", userID),
		})
		return
	}
	user.Blocked = true
	initializer.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("User with ID %s is blocked", userID),
		"user":    user,
	})
}

func UnblockUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	initializer.DB.First(&user, userID)
	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"ERROR": fmt.Sprintf("User with ID %s not found", userID),
		})
		return
	}

	user.Blocked = false
	initializer.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("User with ID %s is unblocked", userID),
		"user":    user,
	})
}

func ListOrders(c *gin.Context) {
	var orders []models.Order
	if err := initializer.DB.Preload("Products").Find(&orders).Error; err != nil {
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
			}
			responseProducts = append(responseProducts, respProduct)
		}

		var user models.User
		if err := initializer.DB.First(&user, order.UserID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve user details",
			})
			return
		}

		respUser := map[string]interface{}{
			"email":    user.Email,
			"Phone no": user.Phoneno,
		}

		var address models.Address
		if err := initializer.DB.First(&address, order.AddressID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve address details",
			})
			return
		}

		fullAddress := fmt.Sprintf("%s, %s, %s, %s, %s",
			address.Street, address.City, address.State, address.Country, address.PostalCode)

		respOrder := map[string]interface{}{
			"user_id":        order.UserID,
			"user_details":   respUser,
			"order_id":       order.ID,
			"address_id":     order.AddressID,
			"address":        fullAddress,
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

func UserOrdersCancel(c *gin.Context) {
	userID := c.Param("id")

	var orders []models.Order
	if err := initializer.DB.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Orders not found for the user",
		})
		return
	}

	var requestBody struct {
		OrderID uint `json:"order_id"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Find the order to cancel
	var orderToCancel models.Order
	found := false
	for _, order := range orders {
		if order.ID == requestBody.OrderID && order.Status != "cancelled" {
			orderToCancel = order
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found or already cancelled",
		})
		return
	}

	// Get user details for email
	var user models.User
	if err := initializer.DB.Where("id = ?", orderToCancel.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user details",
		})
		return
	}

	// Cancel the order
	orderToCancel.Status = "cancelled"
	if err := initializer.DB.Save(&orderToCancel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cancel the order",
		})
		return
	}

	// Prepare the response with canceled order details and user email
	responseCancelledOrder := map[string]interface{}{
		"order_id":     orderToCancel.ID,
		"order_status": orderToCancel.Status,
		"user_email":   user.Email,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Order cancelled successfully",
		"cancelled_order": responseCancelledOrder,
	})
}

func EditOrderStatus(c *gin.Context) {
	userID := c.Param("id")

	var orders []models.Order
	if err := initializer.DB.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Orders not found for the user",
		})
		return
	}

	var requestBody struct {
		OrderID   uint   `json:"order_id"`
		NewStatus string `json:"new_status"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate specific order statuses
	validStatuses := []string{"pending", "processing", "shipped", "delivered"}
	validStatus := false
	for _, status := range validStatuses {
		if requestBody.NewStatus == status {
			validStatus = true
			break
		}
	}

	if !validStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order status",
		})
		return
	}

	// Find the order to edit
	var orderToEdit models.Order
	found := false
	for _, order := range orders {
		if order.ID == requestBody.OrderID {
			orderToEdit = order
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found",
		})
		return
	}

	// Edit the order status
	orderToEdit.Status = requestBody.NewStatus
	if err := initializer.DB.Save(&orderToEdit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to edit the order status",
		})
		return
	}

	// Prepare the response with edited order details
	responseEditedOrder := map[string]interface{}{
		"order_id":   orderToEdit.ID,
		"new_status": orderToEdit.Status,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Order status edited successfully",
		"edited_order": responseEditedOrder,
	})
}
