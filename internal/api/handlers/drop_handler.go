package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"datedrop/internal/api/middleware"
	"datedrop/internal/domain/entities"
	"datedrop/internal/services"
)

type DropHandler struct {
	dropService *services.DropService
}

func NewDropHandler(dropService *services.DropService) *DropHandler {
	return &DropHandler{dropService: dropService}
}

func (h *DropHandler) GetCurrentDrop(c *gin.Context) {
	userID := middleware.GetUserID(c)

	drop, err := h.dropService.GetCurrentDrop(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case services.ErrDropNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "no current drop"})
		case services.ErrDropExpired:
			c.JSON(http.StatusGone, gin.H{"error": "drop has expired"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, drop)
}

func (h *DropHandler) AcceptDrop(c *gin.Context) {
	userID := middleware.GetUserID(c)
	dropID := c.Param("id")

	drop, err := h.dropService.AcceptDrop(c.Request.Context(), userID, dropID)
	if err != nil {
		switch err {
		case services.ErrDropNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case services.ErrDropExpired:
			c.JSON(http.StatusGone, gin.H{"error": err.Error()})
		case services.ErrNotInDrop:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, drop)
}

func (h *DropHandler) DeclineDrop(c *gin.Context) {
	userID := middleware.GetUserID(c)
	dropID := c.Param("id")

	drop, err := h.dropService.DeclineDrop(c.Request.Context(), userID, dropID)
	if err != nil {
		switch err {
		case services.ErrDropNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case services.ErrNotInDrop:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, drop)
}

func (h *DropHandler) GetDropHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)

	drops, err := h.dropService.GetDropHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if drops == nil {
		drops = []*entities.Drop{}
	}

	c.JSON(http.StatusOK, gin.H{"drops": drops})
}
