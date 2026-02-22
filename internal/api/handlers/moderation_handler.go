package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"datedrop/internal/api/middleware"
	"datedrop/internal/services"
)

type ModerationHandler struct {
	moderationService *services.ModerationService
}

func NewModerationHandler(moderationService *services.ModerationService) *ModerationHandler {
	return &ModerationHandler{moderationService: moderationService}
}

func (h *ModerationHandler) BlockUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	blockedID := c.Param("user_id")

	err := h.moderationService.BlockUser(c.Request.Context(), userID, blockedID)
	if err != nil {
		switch err {
		case services.ErrCannotBlockSelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user blocked"})
}

func (h *ModerationHandler) UnblockUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	blockedID := c.Param("user_id")

	err := h.moderationService.UnblockUser(c.Request.Context(), userID, blockedID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user unblocked"})
}

type ReportHTTPRequest struct {
	ReportedID string `json:"reported_id" binding:"required"`
	Category   string `json:"category" binding:"required"`
	Details    string `json:"details"`
}

func (h *ModerationHandler) ReportUser(c *gin.Context) {
	var req ReportHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)

	report, err := h.moderationService.ReportUser(c.Request.Context(), userID, services.ReportRequest{
		ReportedID: req.ReportedID,
		Category:   req.Category,
		Details:    req.Details,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}
