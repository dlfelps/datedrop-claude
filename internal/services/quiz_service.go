package services

import (
	"context"
	"errors"
	"time"

	"datedrop/internal/domain/entities"
	"datedrop/internal/repository/memory"
	"datedrop/pkg/utils"
)

var (
	ErrQuestionNotFound = errors.New("question not found")
	ErrAlreadyAnswered  = errors.New("question already answered")
	ErrInvalidResponse  = errors.New("invalid response for question type")
)

type QuizService struct {
	questionRepo *memory.QuestionRepository
	responseRepo *memory.ResponseRepository
	userRepo     *memory.UserRepository
	totalQuestions int
}

func NewQuizService(
	questionRepo *memory.QuestionRepository,
	responseRepo *memory.ResponseRepository,
	userRepo *memory.UserRepository,
	totalQuestions int,
) *QuizService {
	return &QuizService{
		questionRepo:   questionRepo,
		responseRepo:   responseRepo,
		userRepo:       userRepo,
		totalQuestions: totalQuestions,
	}
}

func (s *QuizService) GetQuestions(ctx context.Context) ([]*entities.Question, error) {
	return s.questionRepo.GetAll(ctx)
}

type SubmitResponseRequest struct {
	QuestionID      string `json:"question_id"`
	ScaleValue      *int   `json:"scale_value,omitempty"`
	ChoiceValue     string `json:"choice_value,omitempty"`
	BooleanValue    *bool  `json:"boolean_value,omitempty"`
	ImportanceScore int    `json:"importance_score"`
}

func (s *QuizService) SubmitResponse(ctx context.Context, userID string, req SubmitResponseRequest) (*entities.QuizResponse, error) {
	// Validate question exists
	question, err := s.questionRepo.GetByID(ctx, req.QuestionID)
	if err != nil {
		return nil, ErrQuestionNotFound
	}

	// Check not already answered
	existing, _ := s.responseRepo.GetByUserAndQuestion(ctx, userID, req.QuestionID)
	if existing != nil {
		return nil, ErrAlreadyAnswered
	}

	// Validate response matches question type
	if err := validateResponse(question, req); err != nil {
		return nil, err
	}

	// Clamp importance
	importance := req.ImportanceScore
	if importance < 1 {
		importance = 3 // default
	}
	if importance > 5 {
		importance = 5
	}

	response := &entities.QuizResponse{
		ID:              utils.GenerateID(),
		UserID:          userID,
		QuestionID:      req.QuestionID,
		ScaleValue:      req.ScaleValue,
		ChoiceValue:     req.ChoiceValue,
		BooleanValue:    req.BooleanValue,
		ImportanceScore: importance,
		CreatedAt:       time.Now(),
	}

	if err := s.responseRepo.Create(ctx, response); err != nil {
		return nil, err
	}

	// Check if quiz is complete
	count, _ := s.responseRepo.CountByUserID(ctx, userID)
	if count >= s.totalQuestions {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err == nil {
			user.QuizCompleted = true
			user.UpdatedAt = time.Now()
			s.userRepo.Update(ctx, user)
		}
	}

	return response, nil
}

type QuizStatus struct {
	UserID         string  `json:"user_id"`
	TotalQuestions int     `json:"total_questions"`
	Answered       int     `json:"answered"`
	Completion     float64 `json:"completion_percentage"`
	QuizCompleted  bool    `json:"quiz_completed"`
}

func (s *QuizService) GetStatus(ctx context.Context, userID string) (*QuizStatus, error) {
	count, _ := s.responseRepo.CountByUserID(ctx, userID)
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	completion := float64(count) / float64(s.totalQuestions) * 100
	if completion > 100 {
		completion = 100
	}

	return &QuizStatus{
		UserID:         userID,
		TotalQuestions: s.totalQuestions,
		Answered:       count,
		Completion:     completion,
		QuizCompleted:  user.QuizCompleted,
	}, nil
}

func validateResponse(q *entities.Question, req SubmitResponseRequest) error {
	switch q.ResponseType {
	case entities.ResponseTypeScale5:
		if req.ScaleValue == nil || *req.ScaleValue < 1 || *req.ScaleValue > 5 {
			return ErrInvalidResponse
		}
	case entities.ResponseTypeScale7:
		if req.ScaleValue == nil || *req.ScaleValue < 1 || *req.ScaleValue > 7 {
			return ErrInvalidResponse
		}
	case entities.ResponseTypeMultipleChoice:
		if req.ChoiceValue == "" {
			return ErrInvalidResponse
		}
		// Validate choice is in options
		found := false
		for _, opt := range q.Options {
			if opt == req.ChoiceValue {
				found = true
				break
			}
		}
		if !found {
			return ErrInvalidResponse
		}
	case entities.ResponseTypeBoolean:
		if req.BooleanValue == nil {
			return ErrInvalidResponse
		}
	}
	return nil
}
