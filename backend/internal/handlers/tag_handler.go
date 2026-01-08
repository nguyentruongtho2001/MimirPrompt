package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"mimirprompt/internal/cache"
	"mimirprompt/internal/models"
	"mimirprompt/internal/repository"

	"github.com/gin-gonic/gin"
)

// TagHandler handles HTTP requests for tags
type TagHandler struct {
	repo  *repository.TagRepository
	db    *sql.DB
	cache cache.CacheService
}

// NewTagHandler creates a new tag handler
func NewTagHandler(repo *repository.TagRepository, db *sql.DB) *TagHandler {
	return &TagHandler{
		repo:  repo,
		db:    db,
		cache: cache.GetCache(),
	}
}

// List returns all tags (with caching)
// GET /api/tags
func (h *TagHandler) List(c *gin.Context) {
	ctx := context.Background()

	// Try to get from cache
	if cached, err := h.cache.Get(ctx, cache.KeyTags); err == nil {
		var tags []models.Tag
		if json.Unmarshal(cached, &tags) == nil {
			c.JSON(http.StatusOK, gin.H{"data": tags})
			return
		}
	}

	// Cache miss - fetch from DB
	tags, err := h.repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store in cache
	h.cache.Set(ctx, cache.KeyTags, tags, cache.TTLTags)

	c.JSON(http.StatusOK, gin.H{"data": tags})
}

// Get returns a single tag by ID
// GET /api/tags/:id
func (h *TagHandler) Get(c *gin.Context) {
	id := c.Param("id")

	tag, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	c.JSON(http.StatusOK, tag)
}

// Create creates a new tag
// POST /api/tags
func (h *TagHandler) Create(c *gin.Context) {
	var input models.TagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tagID, err := h.repo.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	h.cache.Delete(context.Background(), cache.KeyTags)

	c.JSON(http.StatusCreated, gin.H{
		"id":      tagID,
		"name":    input.Name,
		"message": "Tag created successfully",
	})
}

// Update updates an existing tag
// PUT /api/tags/:id
func (h *TagHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var input models.TagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Update(id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Invalidate cache
	h.cache.Delete(context.Background(), cache.KeyTags)

	c.JSON(http.StatusOK, gin.H{"message": "Tag updated successfully"})
}

// Delete deletes a tag
// DELETE /api/tags/:id
func (h *TagHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	affected, err := h.repo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	// Invalidate cache
	h.cache.Delete(context.Background(), cache.KeyTags)

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

// GetStats returns database statistics (with caching)
// GET /api/stats
func (h *TagHandler) GetStats(c *gin.Context) {
	ctx := context.Background()

	// Try to get from cache
	if cached, err := h.cache.Get(ctx, cache.KeyStats); err == nil {
		var stats map[string]interface{}
		if json.Unmarshal(cached, &stats) == nil {
			c.JSON(http.StatusOK, stats)
			return
		}
	}

	// Cache miss - fetch from DB
	stats, err := h.repo.GetStats(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store in cache
	h.cache.Set(ctx, cache.KeyStats, stats, cache.TTLStats)

	c.JSON(http.StatusOK, stats)
}
