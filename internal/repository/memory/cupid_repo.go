package memory

import (
	"context"
	"errors"
	"sync"

	"datedrop/internal/domain/entities"
)

var ErrCupidNotFound = errors.New("cupid nomination not found")

type CupidRepository struct {
	mu          sync.RWMutex
	nominations map[string]*entities.CupidNomination
}

func NewCupidRepository() *CupidRepository {
	return &CupidRepository{
		nominations: make(map[string]*entities.CupidNomination),
	}
}

func (r *CupidRepository) Create(ctx context.Context, nom *entities.CupidNomination) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nominations[nom.ID] = nom
	return nil
}

func (r *CupidRepository) GetByID(ctx context.Context, id string) (*entities.CupidNomination, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	nom, exists := r.nominations[id]
	if !exists {
		return nil, ErrCupidNotFound
	}
	return nom, nil
}

func (r *CupidRepository) Update(ctx context.Context, nom *entities.CupidNomination) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.nominations[nom.ID]; !exists {
		return ErrCupidNotFound
	}
	r.nominations[nom.ID] = nom
	return nil
}

func (r *CupidRepository) GetByUserID(ctx context.Context, userID string) ([]*entities.CupidNomination, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.CupidNomination
	for _, nom := range r.nominations {
		if nom.User1ID == userID || nom.User2ID == userID || nom.NominatorID == userID {
			result = append(result, nom)
		}
	}
	return result, nil
}
