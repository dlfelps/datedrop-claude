package memory

import (
	"context"
	"errors"
	"sync"

	"datedrop/internal/domain/entities"
)

var ErrResponseNotFound = errors.New("response not found")

type ResponseRepository struct {
	mu        sync.RWMutex
	responses map[string]*entities.QuizResponse // keyed by ID
}

func NewResponseRepository() *ResponseRepository {
	return &ResponseRepository{
		responses: make(map[string]*entities.QuizResponse),
	}
}

func (r *ResponseRepository) Create(ctx context.Context, resp *entities.QuizResponse) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.responses[resp.ID] = resp
	return nil
}

func (r *ResponseRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.QuizResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.QuizResponse
	for _, resp := range r.responses {
		if resp.UserID == userID {
			result = append(result, resp)
		}
	}
	return result, nil
}

func (r *ResponseRepository) GetByUserAndQuestion(ctx context.Context, userID, questionID string) (*entities.QuizResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, resp := range r.responses {
		if resp.UserID == userID && resp.QuestionID == questionID {
			return resp, nil
		}
	}
	return nil, ErrResponseNotFound
}

func (r *ResponseRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, resp := range r.responses {
		if resp.UserID == userID {
			count++
		}
	}
	return count, nil
}
