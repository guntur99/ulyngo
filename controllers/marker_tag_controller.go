package controllers

import (
	"net/http"
	"ulyngo/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MarkerTagController struct {
	DB *gorm.DB
}

func NewMarkerTagController(db *gorm.DB) *MarkerTagController {
	return &MarkerTagController{DB: db}
}

// CreateTagInput defines the structure for creating a new marker tag.
type CreateTagInput struct {
	Name string `json:"name" binding:"required"`
}

// CreateTag handles the creation of a new marker tag. (Admin Protected)
func (tc *MarkerTagController) CreateTag(c *gin.Context) {
	var input CreateTagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag := models.MarkerTag{
		Name: input.Name,
	}

	if err := tc.DB.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tag created successfully", "tag": tag})
}

// GetAllTags retrieves all marker tags. (Public)
func (tc *MarkerTagController) GetAllTags(c *gin.Context) {
	var tags []models.MarkerTag
	if err := tc.DB.Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

// GetTagByID retrieves a marker tag by its ID. (Public)
func (tc *MarkerTagController) GetTagByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID format"})
		return
	}

	var tag models.MarkerTag
	if err := tc.DB.First(&tag, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag: " + err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, tag)
}

// UpdateTagInput defines the structure for updating a marker tag.
type UpdateTagInput struct {
	Name *string `json:"name"`
}

// UpdateTag handles the update of an existing marker tag. (Admin Protected)
func (tc *MarkerTagController) UpdateTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID format"})
		return
	}

	var input UpdateTagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tag models.MarkerTag
	if err := tc.DB.First(&tag, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag: " + err.Error()})
		}
		return
	}

	if input.Name != nil {
		tag.Name = *input.Name
	}

	if err := tc.DB.Save(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag updated successfully", "tag": tag})
}

// DeleteTag handles the deletion of a marker tag. (Admin Protected)
func (tc *MarkerTagController) DeleteTag(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID format"})
		return
	}

	// Check if tag exists
	var tag models.MarkerTag
	if err := tc.DB.First(&tag, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tag: " + err.Error()})
		}
		return
	}

	// Perform soft delete
	if err := tc.DB.Delete(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tag: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}
