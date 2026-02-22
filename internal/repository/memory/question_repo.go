package memory

import (
	"context"
	"errors"
	"sync"

	"datedrop/internal/domain/entities"
)

var ErrQuestionNotFound = errors.New("question not found")

type QuestionRepository struct {
	mu        sync.RWMutex
	questions map[string]*entities.Question
}

func NewQuestionRepository() *QuestionRepository {
	return &QuestionRepository{
		questions: make(map[string]*entities.Question),
	}
}

func (r *QuestionRepository) Create(ctx context.Context, q *entities.Question) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.questions[q.ID] = q
	return nil
}

func (r *QuestionRepository) GetByID(ctx context.Context, id string) (*entities.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q, exists := r.questions[id]
	if !exists {
		return nil, ErrQuestionNotFound
	}
	return q, nil
}

func (r *QuestionRepository) GetAll(ctx context.Context) ([]*entities.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entities.Question, 0, len(r.questions))
	for _, q := range r.questions {
		result = append(result, q)
	}
	return result, nil
}

func (r *QuestionRepository) GetByDomain(ctx context.Context, domain entities.QuestionDomain) ([]*entities.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.Question
	for _, q := range r.questions {
		if q.Domain == domain {
			result = append(result, q)
		}
	}
	return result, nil
}

func (r *QuestionRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.questions), nil
}
