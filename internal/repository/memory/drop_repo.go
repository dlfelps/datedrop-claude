package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"datedrop/internal/domain/entities"
)

var ErrDropNotFound = errors.New("drop not found")

type DropRepository struct {
	mu    sync.RWMutex
	drops map[string]*entities.Drop
}

func NewDropRepository() *DropRepository {
	return &DropRepository{
		drops: make(map[string]*entities.Drop),
	}
}

func (r *DropRepository) Create(ctx context.Context, drop *entities.Drop) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drops[drop.ID] = drop
	return nil
}

func (r *DropRepository) GetByID(ctx context.Context, id string) (*entities.Drop, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	drop, exists := r.drops[id]
	if !exists {
		return nil, ErrDropNotFound
	}
	return drop, nil
}

func (r *DropRepository) Update(ctx context.Context, drop *entities.Drop) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.drops[drop.ID]; !exists {
		return ErrDropNotFound
	}
	r.drops[drop.ID] = drop
	return nil
}

func (r *DropRepository) GetCurrentByUserID(ctx context.Context, userID string) (*entities.Drop, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, drop := range r.drops {
		if (drop.User1ID == userID || drop.User2ID == userID) &&
			drop.Status != entities.DropStatusCompleted &&
			drop.Status != entities.DropStatusArchived &&
			drop.Status != entities.DropStatusExpired &&
			drop.Status != entities.DropStatusDeclined {
			return drop, nil
		}
	}
	return nil, ErrDropNotFound
}

func (r *DropRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.Drop, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.Drop
	for _, drop := range r.drops {
		if drop.User1ID == userID || drop.User2ID == userID {
			result = append(result, drop)
		}
	}
	return result, nil
}

func (r *DropRepository) GetMatchedPairs(ctx context.Context, withinWeeks int) ([][2]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cutoff := time.Now().Add(-time.Duration(withinWeeks) * 7 * 24 * time.Hour)
	var pairs [][2]string
	for _, drop := range r.drops {
		if drop.CreatedAt.After(cutoff) {
			pairs = append(pairs, [2]string{drop.User1ID, drop.User2ID})
		}
	}
	return pairs, nil
}
