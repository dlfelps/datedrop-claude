package services

import (
	"context"
	"errors"

	"datedrop/internal/domain/entities"
	"datedrop/internal/repository/memory"
)

var (
	ErrDropNotFound  = errors.New("drop not found")
	ErrDropExpired   = errors.New("drop has expired")
	ErrNotInDrop     = errors.New("user is not part of this drop")
)

type DropService struct {
	dropRepo     *memory.DropRepository
	notifService *NotificationService
}

func NewDropService(dropRepo *memory.DropRepository, notifService *NotificationService) *DropService {
	return &DropService{
		dropRepo:     dropRepo,
		notifService: notifService,
	}
}

func (s *DropService) GetCurrentDrop(ctx context.Context, userID string) (*entities.Drop, error) {
	drop, err := s.dropRepo.GetCurrentByUserID(ctx, userID)
	if err != nil {
		return nil, ErrDropNotFound
	}

	// Check expiration
	if drop.IsExpired() && drop.Status != entities.DropStatusMatched && drop.Status != entities.DropStatusCompleted {
		drop.TransitionTo(entities.DropStatusExpired)
		s.dropRepo.Update(ctx, drop)
		return nil, ErrDropExpired
	}

	return drop, nil
}

func (s *DropService) AcceptDrop(ctx context.Context, userID, dropID string) (*entities.Drop, error) {
	drop, err := s.dropRepo.GetByID(ctx, dropID)
	if err != nil {
		return nil, ErrDropNotFound
	}

	if drop.IsExpired() {
		drop.TransitionTo(entities.DropStatusExpired)
		s.dropRepo.Update(ctx, drop)
		return nil, ErrDropExpired
	}

	if err := drop.Accept(userID); err != nil {
		return nil, ErrNotInDrop
	}

	if err := s.dropRepo.Update(ctx, drop); err != nil {
		return nil, err
	}

	// Notify on match
	if drop.Status == entities.DropStatusMatched {
		s.notifService.NotifyMatch(drop.User1ID, drop.User2ID, drop.ID)
	}

	return drop, nil
}

func (s *DropService) DeclineDrop(ctx context.Context, userID, dropID string) (*entities.Drop, error) {
	drop, err := s.dropRepo.GetByID(ctx, dropID)
	if err != nil {
		return nil, ErrDropNotFound
	}

	if err := drop.Decline(userID); err != nil {
		return nil, ErrNotInDrop
	}

	if err := s.dropRepo.Update(ctx, drop); err != nil {
		return nil, err
	}

	return drop, nil
}

func (s *DropService) GetDropHistory(ctx context.Context, userID string) ([]*entities.Drop, error) {
	return s.dropRepo.GetByUserID(ctx, userID)
}
