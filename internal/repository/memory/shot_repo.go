package memory

import (
	"context"
	"errors"
	"sync"

	"datedrop/internal/domain/entities"
)

var ErrShotNotFound = errors.New("shot not found")

type ShotRepository struct {
	mu    sync.RWMutex
	shots map[string]*entities.Shot
}

func NewShotRepository() *ShotRepository {
	return &ShotRepository{
		shots: make(map[string]*entities.Shot),
	}
}

func (r *ShotRepository) Create(ctx context.Context, shot *entities.Shot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shots[shot.ID] = shot
	return nil
}

func (r *ShotRepository) GetByShooterAndTarget(ctx context.Context, shooterID, targetID string) (*entities.Shot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, shot := range r.shots {
		if shot.ShooterID == shooterID && shot.TargetID == targetID {
			return shot, nil
		}
	}
	return nil, ErrShotNotFound
}

func (r *ShotRepository) GetMutualShots(ctx context.Context, userID string) ([]*entities.Shot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find shots where user is the target
	targetsOfUser := make(map[string]bool)
	for _, shot := range r.shots {
		if shot.ShooterID == userID {
			targetsOfUser[shot.TargetID] = true
		}
	}

	// Find reverse shots (target shot back at user)
	var mutual []*entities.Shot
	for _, shot := range r.shots {
		if shot.TargetID == userID && targetsOfUser[shot.ShooterID] {
			mutual = append(mutual, shot)
		}
	}
	return mutual, nil
}
