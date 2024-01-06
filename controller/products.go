package controller

import (
	"ecommerce-mobilestore/initializer"
	"ecommerce-mobilestore/models"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func AddProduct(c *gin.Context) {
	ok := adminLoggedStatus
	if ok {
		var product models.Product

		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
			return
		}

		var category models.Category
		result := initializer.DB.First(&category, "id = ?", product.CategoryID)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrive category"})
			return
		}

		if result := initializer.DB.Create(&product); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		product.Available = true
		initializer.DB.Save(&product)

		c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully", "product": product})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "needs to login",
		})
	}
}

func EditProduct(c *gin.Context) {
	productID := c.Param("id")

	var product models.Product
	if err := initializer.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var updatedProduct models.Product
	if err := c.ShouldBindJSON(&updatedProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	product.Name = updatedProduct.Name
	product.Price = updatedProduct.Price
	product.Quantity = updatedProduct.Quantity
	product.CategoryID = updatedProduct.CategoryID

	if result := initializer.DB.Save(&product); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully", "product": product})
}

func DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	var product models.Product
	if err := initializer.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	product.Available = false
	
	if err := initializer.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't Delete Product"})
		return
	}

	
	if result := initializer.DB.Save(&product); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product soft deleted successfully", "product": product})
}

func UploadProductImage(c *gin.Context) {
	// Extract product ID from the URL parameters or request body
	productID := c.Param("id")

	// Validate and sanitize productID as needed
	var product models.Product
	if err := initializer.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	// Limit 10mb
	c.Request.ParseMultipartForm(10 * 1024 * 1024)

	files := c.Request.MultipartForm.File["product_image"]

	var Response []map[string]interface{}
	for _, file := range files {
		// Save file to server
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to open file",
			})
			return
		}

		tempFile, err := os.CreateTemp("uploads", "upload-*.jpg")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create temporary file",
			})
			return
		}

		fileBytes, err := io.ReadAll(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read file",
			})
			return
		}

		var productimg = models.Productimg{
			ProductID: product.ID,
			Image: file.Filename,
		}
		if result := initializer.DB.Create(&productimg); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add image"})
			return
		}

		tempFile.Write(fileBytes)
		response := map[string]interface{}{
			"File name": file.Filename,
		}

		Response = append(Response, response)
		defer tempFile.Close()

	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Images uploaded successfully",
		"images":       Response,
	})

}
