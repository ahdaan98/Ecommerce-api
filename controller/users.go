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
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
	"golang.org/x/crypto/bcrypt"
)

const (
	TWILIO_ACCOUNT_SID = "ACe1c11ec1cb089d28198af267808c8c1d"
	TWILIO_AUTH_TOKEN  = "a20cd65e286ce7ebd7f325f973738186"
	VERIFY_SERVICE_SID = "VAbcc39110308b0f8e592574f97baa102e"
)

var client *twilio.RestClient = twilio.NewRestClientWithParams(twilio.ClientParams{
	Username: TWILIO_ACCOUNT_SID,
	Password: TWILIO_AUTH_TOKEN,
})

var phonenumber string
var userDetails map[string]string
var userLoggedStatus = false

func UserLogged(c *gin.Context) {
	userLoggedStatus = true
}

func SignUp(c *gin.Context) {
	// Get the email/pass off req body
	var body struct {
		Email    string
		Phoneno  string
		Password string
	}

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	if len(body.Password) < 8 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Password must contain 8 Letters",
		})
		return
	}

	var user models.User
	initializer.DB.First(&user, "email = ?", body.Email)
	if user.Email != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User is already existed",
		})
		return
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	params := &openapi.CreateVerificationParams{}
	params.SetTo(body.Phoneno)
	params.SetChannel("sms")

	resp, err := client.VerifyV2.CreateVerification(VERIFY_SERVICE_SID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_=resp

	userDetails = map[string]string{
		"email":    body.Email,
		"phoneno":  body.Phoneno,
		"password": string(hash),
	}

	// respond
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Verification code sent",
		"phonenumber": body.Phoneno,
	})

	phonenumber = body.Phoneno
}

func VerifyOtpAndSignupUser(c *gin.Context) {
	userDetailsRaw := userDetails

	var requestData struct {
		Code string `json:"code"`
	}

	number := userDetailsRaw["phoneno"]
	if number == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Phone number not found",
		})
		return
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(requestData)
	fmt.Println(requestData.Code)

	params := &openapi.CreateVerificationCheckParams{}
	params.SetTo(number)
	params.SetCode(requestData.Code)

	resp, err := client.VerifyV2.CreateVerificationCheck(VERIFY_SERVICE_SID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create JSON response
	response := gin.H{"status": *resp.Status}
	if *resp.Status == "approved" {
		response["details"] = gin.H{"sid": *resp.Sid}
	}

	SigningUP(c, userDetailsRaw)

	c.JSON(http.StatusOK, response)
}

func SigningUP(c *gin.Context, userDetails map[string]string) {
	// Use GORM to create the user
	user := models.User{
		Email:    userDetails["email"],
		Phoneno:  userDetails["phoneno"],
		Password: userDetails["password"],
		Blocked:  false,
	}

	result := initializer.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	response:=map[string]interface{}{
		"email":user.Email,
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Lezuse",
		"user":    response["email"],
	})
}

func LoginUser(c *gin.Context) {
	// Get the email and pass off req body
	var body struct {
		Email    string
		Password string
	}

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Look up requested user
	var user models.User
	initializer.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// Compare sent in pass with saved user pass hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// generate a jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
		return
	}

	// send it
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("usertoken", tokenString, 3600*24*30, "", "", false, true)
	userLoggedStatus = true
	c.JSON(http.StatusOK, gin.H{
		"message": "Succesdfully Login",
	})
}

func LogoutUser(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("usertoken", "", -1, "", "", false, true)
	userLoggedStatus = false
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully Log out",
	})
}

func ListProducts(c *gin.Context) {
	ok := userLoggedStatus
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized access",
		})
		return
	}
	var products []models.Product
	if err := initializer.DB.Where("available = ?", true).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve products",
		})
		return
	}

	// Prepare the response with product details and image names
	var responseProducts []map[string]interface{}
	for _, product := range products {
		var productImages []models.Productimg
		if err := initializer.DB.Where("product_id = ?", product.ID).Find(&productImages).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve product images",
			})
			return
		}

		var imageNames []string
		for _, img := range productImages {
			imageNames = append(imageNames, img.Image)
		}

		responseProduct := map[string]interface{}{
			"id":       product.ID,
			"name":     product.Name,
			"price":    product.Price,
			"quantity": product.Quantity,
			"images":   imageNames,
		}

		responseProducts = append(responseProducts, responseProduct)
	}

	c.JSON(http.StatusOK, gin.H{
		"products": responseProducts,
	})

}

func AddToCart(c *gin.Context) {
	// Get product ID from request parameters
	productID := c.Param("id")

	// Fetch the product
	var product models.Product
	if err := initializer.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve product details",
		})
		return
	}

	// Retrieve user information from context
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

	// Check if the product already exists in the user's cart
	var existingCart models.Cart
	if err := initializer.DB.Where("user_id = ? AND product_id = ?", user.ID, product.ID).First(&existingCart).Error; err == nil {
		// Increment quantity if the product is already in the cart
		existingCart.Quantity++
		if err := initializer.DB.Save(&existingCart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update cart",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Quantity updated in cart",
		})
		return
	}

	// Add the product to the cart with quantity 1
	cart := models.Cart{
		UserID:    user.ID,
		ProductID: product.ID,
		Quantity:  1,
	}
	if err := initializer.DB.Create(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add product to cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product added to cart",
	})
}

func ListCart(c *gin.Context) {
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

	var cartItems []models.Cart
	if err := initializer.DB.Where("user_id = ?", user.ID).Find(&cartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve cart items",
		})
		return
	}

	// Variables to store total quantity and total amount
	totalQuantity := 0
	totalAmount := 0.0

	// Retrieve associated product details with images for each cart item
	var responseCartItems []map[string]interface{}
	for _, cart := range cartItems {
		var product models.Product
		if err := initializer.DB.First(&product, cart.ProductID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve product details",
			})
			return
		}

		// Retrieve associated product images
		var productImages []models.Productimg
		if err := initializer.DB.Where("product_id = ?", product.ID).Find(&productImages).Error; err != nil {
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

		// Build the response for each cart item
		responseCartItem := map[string]interface{}{
			"product_id":    product.ID,
			"product_name":  product.Name,
			"quantity":      cart.Quantity,
			"price":         product.Price,
			"total_amount":  float64(cart.Quantity) * product.Price,
			"product_images": imageNames,
			// Add other product fields as needed
		}

		// Add to the response cart items list
		responseCartItems = append(responseCartItems, responseCartItem)

		// Update total quantity and amount
		totalQuantity += int(cart.Quantity)
		totalAmount += float64(cart.Quantity) * product.Price
	}

	c.JSON(http.StatusOK, gin.H{
		"CART ITEMS":     responseCartItems,
		"TOTAL QUANTITY": totalQuantity,
		"TOTAL AMOUNT":   totalAmount,
	})
}

func RemoveFromCart(c *gin.Context) {
	// Get product ID from request parameters
	productID := c.Param("id")

	// Retrieve user information from context
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

	// Find the cart entry for the product and user
	var cart models.Cart
	if err := initializer.DB.Where("user_id = ? AND product_id = ?", user.ID, productID).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found in cart",
		})
		return
	}

	// Delete the cart entry
	if err := initializer.DB.Delete(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove product from cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product removed from cart",
	})
}

func CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
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

	var order models.Order
	if err := initializer.DB.Where("user_id = ? AND id = ?", user.ID, orderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Order not found in cart",
		})
		return
	}

	order.Status = "cancelled"
	if err := initializer.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cancel the order",
		})
		return
	}

	orders := map[string]interface{}{
		"order id":     order.ID,
		"order status": order.Status,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
		"order":   orders,
	})
}
