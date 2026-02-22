package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"datedrop/internal/api/middleware"
	"datedrop/internal/services"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type CreateUserHTTPRequest struct {
	Email        string   `json:"email" binding:"required"`
	Name         string   `json:"name" binding:"required"`
	DateOfBirth  string   `json:"date_of_birth" binding:"required"`
	Gender       string   `json:"gender" binding:"required"`
	Orientations []string `json:"orientations" binding:"required"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dob, err := parseDate(req.DateOfBirth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_of_birth format, use YYYY-MM-DD"})
		return
	}

	orientations := parseOrientations(req.Orientations)

	user, err := h.userService.CreateUser(c.Request.Context(), services.CreateUserRequest{
		Email:        req.Email,
		Name:         req.Name,
		DateOfBirth:  dob,
		Gender:       parseGender(req.Gender),
		Orientations: orientations,
	})
	if err != nil {
		switch err {
		case services.ErrInvalidEmail:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case services.ErrTooYoung:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case services.ErrEmailTaken:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

type LoginRequest struct {
	Email string `json:"email" binding:"required"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Login(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Mock auth: return user ID as the token
	c.JSON(http.StatusOK, gin.H{
		"token": user.ID,
		"user":  user,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	authUserID := middleware.GetUserID(c)

	if userID != authUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
