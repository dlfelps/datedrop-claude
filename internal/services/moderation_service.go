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
	ErrCannotBlockSelf = errors.New("cannot block yourself")
	ErrBlockNotFound   = errors.New("block not found")
)

type ModerationService struct {
	moderationRepo *memory.ModerationRepository
	notifService   *NotificationService
}

func NewModerationService(moderationRepo *memory.ModerationRepository, notifService *NotificationService) *ModerationService {
	return &ModerationService{
		moderationRepo: moderationRepo,
		notifService:   notifService,
	}
}

func (s *ModerationService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	if blockerID == blockedID {
		return ErrCannotBlockSelf
	}

	block := &entities.Block{
		ID:        utils.GenerateID(),
		BlockerID: blockerID,
		BlockedID: blockedID,
		CreatedAt: time.Now(),
	}

	if err := s.moderationRepo.CreateBlock(ctx, block); err != nil {
		return err
	}

	s.notifService.NotifyBlock(blockerID, blockedID)
	return nil
}

func (s *ModerationService) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	return s.moderationRepo.RemoveBlock(ctx, blockerID, blockedID)
}

type ReportRequest struct {
	ReportedID string `json:"reported_id"`
	Category   string `json:"category"`
	Details    string `json:"details"`
}

func (s *ModerationService) ReportUser(ctx context.Context, reporterID string, req ReportRequest) (*entities.Report, error) {
	report := &entities.Report{
		ID:         utils.GenerateID(),
		ReporterID: reporterID,
		ReportedID: req.ReportedID,
		Category:   entities.ReportCategory(req.Category),
		Details:    req.Details,
		Status:     entities.ReportStatusOpen,
		CreatedAt:  time.Now(),
	}

	if err := s.moderationRepo.CreateReport(ctx, report); err != nil {
		return nil, err
	}

	s.notifService.NotifyReport(reporterID, req.ReportedID, req.Category)
	return report, nil
}
