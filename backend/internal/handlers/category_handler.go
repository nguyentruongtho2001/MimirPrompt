package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"mimirprompt/internal/cache"
	"mimirprompt/internal/models"
	"mimirprompt/internal/repository"

	"github.com/gin-gonic/gin"
)

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	repo  *repository.CategoryRepository
	cache cache.CacheService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(repo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		repo:  repo,
		cache: cache.GetCache(),
	}
}

// List returns all categories (with caching)
// GET /api/categories
func (h *CategoryHandler) List(c *gin.Context) {
	ctx := context.Background()

	// Try to get from cache
	if cached, err := h.cache.Get(ctx, cache.KeyCategories); err == nil {
		var categories []models.Category
		if json.Unmarshal(cached, &categories) == nil {
			c.JSON(http.StatusOK, categories)
			return
		}
	}

	// Cache miss - fetch from DB
	categories, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store in cache
	h.cache.Set(ctx, cache.KeyCategories, categories, cache.TTLCategories)

	c.JSON(http.StatusOK, categories)
}

// Get returns a single category by ID
// GET /api/categories/:id
func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	category, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// Create creates a new category
// POST /api/categories
func (h *CategoryHandler) Create(c *gin.Context) {
	var input models.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	categoryID, err := h.repo.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	h.cache.Delete(context.Background(), cache.KeyCategories)

	c.JSON(http.StatusCreated, gin.H{
		"id":      categoryID,
		"message": "Category created successfully",
	})
}

// Update updates an existing category
// PUT /api/categories/:id
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input models.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Update(id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	h.cache.Delete(context.Background(), cache.KeyCategories)

	c.JSON(http.StatusOK, gin.H{"message": "Category updated successfully"})
}

// Delete deletes a category
// DELETE /api/categories/:id
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	h.cache.Delete(context.Background(), cache.KeyCategories)

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
