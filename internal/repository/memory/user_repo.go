package memory

import (
	"context"
	"errors"
	"strings"
	"sync"

	"datedrop/internal/domain/entities"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*entities.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*entities.User),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if strings.EqualFold(user.Email, email) {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return ErrUserNotFound
	}
	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) GetActiveQuizComplete(ctx context.Context, excludeIDs []string) ([]*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	excludeSet := make(map[string]bool, len(excludeIDs))
	for _, id := range excludeIDs {
		excludeSet[id] = true
	}

	var result []*entities.User
	for _, user := range r.users {
		if user.IsActive && user.QuizCompleted && !excludeSet[user.ID] {
			result = append(result, user)
		}
	}
	return result, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.User
	for _, user := range r.users {
		result = append(result, user)
	}
	return result, nil
}
