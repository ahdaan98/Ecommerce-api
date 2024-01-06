package controller

import (
	"ecommerce-mobilestore/initializer"
	"ecommerce-mobilestore/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBind(&category); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
			return
	}

	if result := initializer.DB.Create(&category); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
			return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Category created successfully", "category": category})
}

func UpdateCategory(c *gin.Context) {
	categoryID := c.Param("id")

	var category models.Category
	if err := initializer.DB.First(&category, categoryID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
	}

	var updatedCategory models.Category
	if err := c.ShouldBindJSON(&updatedCategory); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
			return
	}

	category.Name = updatedCategory.Name

	if result := initializer.DB.Save(&category); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
			return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category updated successfully", "category": category})
}

func DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")

	var category models.Category
	if err := initializer.DB.First(&category, categoryID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
	}

	if result := initializer.DB.Delete(&category).Error; result != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
			return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

func GetAllCategories(c *gin.Context) {
	var categories []models.Category
	initializer.DB.Preload("Products").Find(&categories)

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}