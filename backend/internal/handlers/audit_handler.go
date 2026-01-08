package handlers

import (
	"net/http"
	"strconv"

	"mimirprompt/internal/repository"

	"github.com/gin-gonic/gin"
)

// AuditHandler handles HTTP requests for audit logs
type AuditHandler struct {
	repo *repository.AuditRepository
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(repo *repository.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

// List returns a paginated list of audit logs
// GET /api/audit-logs?page=1&per_page=20&entity_type=prompt&action=UPDATE
func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	entityType := c.Query("entity_type")
	action := c.Query("action")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	logs, total, err := h.repo.List(page, perPage, entityType, action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + perPage - 1) / perPage

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	})
}

// GetByEntity returns audit logs for a specific entity
// GET /api/audit-logs/:entity_type/:entity_id
func (h *AuditHandler) GetByEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityID, err := strconv.Atoi(c.Param("entity_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
		return
	}

	logs, err := h.repo.GetByEntityID(entityType, entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}
