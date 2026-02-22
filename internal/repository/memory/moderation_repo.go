package memory

import (
	"context"
	"errors"
	"sync"

	"datedrop/internal/domain/entities"
)

var ErrBlockNotFound = errors.New("block not found")

type ModerationRepository struct {
	mu      sync.RWMutex
	blocks  map[string]*entities.Block  // keyed by "blockerID:blockedID"
	reports map[string]*entities.Report // keyed by ID
}

func NewModerationRepository() *ModerationRepository {
	return &ModerationRepository{
		blocks:  make(map[string]*entities.Block),
		reports: make(map[string]*entities.Report),
	}
}

func blockKey(blockerID, blockedID string) string {
	return blockerID + ":" + blockedID
}

func (r *ModerationRepository) CreateBlock(ctx context.Context, block *entities.Block) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.blocks[blockKey(block.BlockerID, block.BlockedID)] = block
	return nil
}

func (r *ModerationRepository) RemoveBlock(ctx context.Context, blockerID, blockedID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := blockKey(blockerID, blockedID)
	if _, exists := r.blocks[key]; !exists {
		return ErrBlockNotFound
	}
	delete(r.blocks, key)
	return nil
}

func (r *ModerationRepository) IsBlocked(ctx context.Context, userID1, userID2 string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, blocked1 := r.blocks[blockKey(userID1, userID2)]
	_, blocked2 := r.blocks[blockKey(userID2, userID1)]
	return blocked1 || blocked2, nil
}

func (r *ModerationRepository) GetBlockedIDs(ctx context.Context, userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []string
	for _, block := range r.blocks {
		if block.BlockerID == userID {
			ids = append(ids, block.BlockedID)
		}
		if block.BlockedID == userID {
			ids = append(ids, block.BlockerID)
		}
	}
	return ids, nil
}

func (r *ModerationRepository) CreateReport(ctx context.Context, report *entities.Report) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports[report.ID] = report
	return nil
}

func (r *ModerationRepository) GetReports(ctx context.Context) ([]*entities.Report, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*entities.Report
	for _, report := range r.reports {
		result = append(result, report)
	}
	return result, nil
}
