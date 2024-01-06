package middleware

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

func RequireAuthAdmin(c *gin.Context) {
	// Get the cookie off req
	tokenString, err := c.Cookie("admintoken")

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	// Decode/validate it
	// Parse takes the token string and a function for looking up the key. The latter is especially
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}  

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// check the exp
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// Find the admin with token sub
		var admin models.Admin
		initializer.DB.First(&admin, claims["sub"])

		if admin.ID == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// Attach to req
		c.Set("admin", admin)

		// Continue

		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

}
