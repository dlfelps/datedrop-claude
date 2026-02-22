package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"datedrop/internal/services"
)

type MatchingHandler struct {
	matchingService *services.MatchingService
}

func NewMatchingHandler(matchingService *services.MatchingService) *MatchingHandler {
	return &MatchingHandler{matchingService: matchingService}
}

func (h *MatchingHandler) RunMatching(c *gin.Context) {
	drops, err := h.matchingService.RunWeeklyMatching(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches_created": len(drops),
		"drops":           drops,
	})
}
