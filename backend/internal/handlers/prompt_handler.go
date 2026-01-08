package handlers

import (
	"net/http"
	"strconv"

	"mimirprompt/internal/models"
	"mimirprompt/internal/repository"

	"github.com/gin-gonic/gin"
)

// PromptHandler handles HTTP requests for prompts
type PromptHandler struct {
	repo *repository.PromptRepository
}

// NewPromptHandler creates a new prompt handler
func NewPromptHandler(repo *repository.PromptRepository) *PromptHandler {
	return &PromptHandler{repo: repo}
}

// List returns a paginated list of prompts
// GET /api/prompts?page=1&per_page=20
func (h *PromptHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	showHidden := c.Query("show_hidden") == "true"

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	prompts, total, err := h.repo.List(page, perPage, showHidden)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + perPage - 1) / perPage

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       prompts,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// Get returns a single prompt by ID
// GET /api/prompts/:id
func (h *PromptHandler) Get(c *gin.Context) {
	id := c.Param("id")

	prompt, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
		return
	}

	c.JSON(http.StatusOK, prompt)
}

// Create creates a new prompt
// POST /api/prompts
func (h *PromptHandler) Create(c *gin.Context) {
	var input models.PromptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promptID, err := h.repo.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      promptID,
		"message": "Prompt created successfully",
	})
}

// Update updates an existing prompt
// PUT /api/prompts/:id
func (h *PromptHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var input models.PromptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Update(id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prompt updated successfully"})
}

// Delete deletes a prompt
// DELETE /api/prompts/:id
func (h *PromptHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prompt deleted successfully"})
}

// Search searches prompts by query
// GET /api/search?q=...&page=1&per_page=20
func (h *PromptHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	prompts, total, err := h.repo.Search(query, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + perPage - 1) / perPage

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       prompts,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// ListByTag returns prompts filtered by tag
// GET /api/tags/slug/:slug/prompts
func (h *PromptHandler) ListByTag(c *gin.Context) {
	slug := c.Param("slug")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	prompts, total, err := h.repo.ListByTag(slug, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + perPage - 1) / perPage

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       prompts,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// BulkDelete deletes multiple prompts
// POST /api/prompts/bulk-delete
func (h *PromptHandler) BulkDelete(c *gin.Context) {
	var input models.BulkActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No IDs provided"})
		return
	}

	affected, err := h.repo.BulkDelete(input.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Prompts deleted successfully",
		"affected": affected,
	})
}

// BulkUpdateVisibility updates visibility for multiple prompts
// POST /api/prompts/bulk-visibility
func (h *PromptHandler) BulkUpdateVisibility(c *gin.Context) {
	var input models.BulkActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No IDs provided"})
		return
	}

	if input.IsHidden == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_hidden value is required"})
		return
	}

	affected, err := h.repo.BulkUpdateVisibility(input.IDs, *input.IsHidden)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Visibility updated successfully",
		"affected": affected,
	})
}
