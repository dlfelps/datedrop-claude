package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"datedrop/internal/api/middleware"
	"datedrop/internal/services"
)

type SocialHandler struct {
	socialService *services.SocialService
}

func NewSocialHandler(socialService *services.SocialService) *SocialHandler {
	return &SocialHandler{socialService: socialService}
}

type ShootRequest struct {
	TargetID string `json:"target_id" binding:"required"`
}

func (h *SocialHandler) ShootYourShot(c *gin.Context) {
	var req ShootRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)

	shot, mutual, err := h.socialService.ShootYourShot(c.Request.Context(), userID, req.TargetID)
	if err != nil {
		switch err {
		case services.ErrCannotShootSelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case services.ErrShotAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"shot":   shot,
		"mutual": mutual,
	})
}

func (h *SocialHandler) GetMutualShots(c *gin.Context) {
	userID := middleware.GetUserID(c)

	shots, err := h.socialService.GetMutualShots(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mutual_shots": shots})
}

func (h *SocialHandler) BrowseUsers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, err := h.socialService.BrowseUsers(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"page":  page,
	})
}

type CupidNominateHTTPRequest struct {
	User1ID string `json:"user1_id" binding:"required"`
	User2ID string `json:"user2_id" binding:"required"`
}

func (h *SocialHandler) NominateCupid(c *gin.Context) {
	var req CupidNominateHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)

	nom, err := h.socialService.NominateCupid(c.Request.Context(), userID, services.CupidNominateRequest{
		User1ID: req.User1ID,
		User2ID: req.User2ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, nom)
}

func (h *SocialHandler) AcceptCupid(c *gin.Context) {
	userID := middleware.GetUserID(c)
	nomID := c.Param("id")

	nom, err := h.socialService.AcceptCupid(c.Request.Context(), userID, nomID)
	if err != nil {
		switch err {
		case services.ErrCupidNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case services.ErrNotInCupid:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, nom)
}

func (h *SocialHandler) DeclineCupid(c *gin.Context) {
	userID := middleware.GetUserID(c)
	nomID := c.Param("id")

	nom, err := h.socialService.DeclineCupid(c.Request.Context(), userID, nomID)
	if err != nil {
		switch err {
		case services.ErrCupidNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case services.ErrNotInCupid:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, nom)
}
