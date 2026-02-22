package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"datedrop/internal/api/middleware"
	"datedrop/internal/services"
)

type QuizHandler struct {
	quizService *services.QuizService
}

func NewQuizHandler(quizService *services.QuizService) *QuizHandler {
	return &QuizHandler{quizService: quizService}
}

func (h *QuizHandler) GetQuestions(c *gin.Context) {
	questions, err := h.quizService.GetQuestions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": questions})
}

type SubmitResponseHTTPRequest struct {
	QuestionID      string `json:"question_id" binding:"required"`
	ScaleValue      *int   `json:"scale_value,omitempty"`
	ChoiceValue     string `json:"choice_value,omitempty"`
	BooleanValue    *bool  `json:"boolean_value,omitempty"`
	ImportanceScore int    `json:"importance_score"`
}

func (h *QuizHandler) SubmitResponse(c *gin.Context) {
	var req SubmitResponseHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)

	response, err := h.quizService.SubmitResponse(c.Request.Context(), userID, services.SubmitResponseRequest{
		QuestionID:      req.QuestionID,
		ScaleValue:      req.ScaleValue,
		ChoiceValue:     req.ChoiceValue,
		BooleanValue:    req.BooleanValue,
		ImportanceScore: req.ImportanceScore,
	})
	if err != nil {
		switch err {
		case services.ErrQuestionNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case services.ErrAlreadyAnswered:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case services.ErrInvalidResponse:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *QuizHandler) GetStatus(c *gin.Context) {
	userID := c.Param("user_id")

	status, err := h.quizService.GetStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}
