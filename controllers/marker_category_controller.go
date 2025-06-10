package controllers

import (
	"net/http"
	"ulyngo/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarkerCategoryController struct {
	DB *gorm.DB
}

func NewMarkerCategoryController(db *gorm.DB) *MarkerCategoryController {
	return &MarkerCategoryController{DB: db}
}

// CreateCategoryInput defines the structure for creating a new marker category.
type CreateCategoryInput struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

// CreateCategory handles the creation of a new marker category. (Admin Protected)
func (cc *MarkerCategoryController) CreateCategory(c *gin.Context) {
	var input CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := models.MarkerCategory{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := cc.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Category created successfully", "category": category})
}

// GetAllCategories retrieves all marker categories. (Public)
func (cc *MarkerCategoryController) GetAllCategories(c *gin.Context) {
	var categories []models.MarkerCategory
	if err := cc.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetCategoryByID retrieves a marker category by its ID. (Public)
func (cc *MarkerCategoryController) GetCategoryByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID format"})
		return
	}

	var category models.MarkerCategory
	if err := cc.DB.First(&category, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, category)
}

// UpdateCategoryInput defines the structure for updating a marker category.
type UpdateCategoryInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// UpdateCategory handles the update of an existing marker category. (Admin Protected)
func (cc *MarkerCategoryController) UpdateCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID format"})
		return
	}

	var input UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var category models.MarkerCategory
	if err := cc.DB.First(&category, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category: " + err.Error()})
		}
		return
	}

	if input.Name != nil {
		category.Name = *input.Name
	}
	if input.Description != nil {
		category.Description = input.Description
	}

	if err := cc.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category updated successfully", "category": category})
}

// DeleteCategory handles the deletion of a marker category. (Admin Protected)
func (cc *MarkerCategoryController) DeleteCategory(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID format"})
		return
	}

	// Check if category exists
	var category models.MarkerCategory
	if err := cc.DB.First(&category, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category: " + err.Error()})
		}
		return
	}

	// Perform soft delete
	if err := cc.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
