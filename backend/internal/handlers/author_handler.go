package handlers

import (
	"net/http"
	"strconv"

	"mimirprompt/internal/models"
	"mimirprompt/internal/repository"

	"github.com/gin-gonic/gin"
)

type AuthorHandler struct {
	repo *repository.AuthorRepository
}

func NewAuthorHandler(repo *repository.AuthorRepository) *AuthorHandler {
	return &AuthorHandler{repo: repo}
}

// List returns all authors
func (h *AuthorHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}

	// If search query provided
	if search != "" {
		authors, err := h.repo.Search(search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, authors)
		return
	}

	authors, total, err := h.repo.List(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + perPage - 1) / perPage

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       authors,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

// Get returns an author by ID
func (h *AuthorHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	author, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Author not found"})
		return
	}

	c.JSON(http.StatusOK, author)
}

// Create creates a new author
func (h *AuthorHandler) Create(c *gin.Context) {
	var input models.AuthorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default platform to twitter if not specified
	if input.Platform == "" {
		input.Platform = "twitter"
	}

	id, err := h.repo.Create(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	author, _ := h.repo.GetByID(id)
	c.JSON(http.StatusCreated, author)
}

// Update updates an existing author
func (h *AuthorHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	var input models.AuthorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Update(id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	author, _ := h.repo.GetByID(id)
	c.JSON(http.StatusOK, author)
}

// Delete deletes an author
func (h *AuthorHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Author deleted successfully"})
}
