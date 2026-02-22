package services

import (
	"context"
	"errors"
	"time"

	"datedrop/internal/config"
	"datedrop/internal/domain/entities"
	"datedrop/internal/repository/memory"
	"datedrop/pkg/utils"
)

var (
	ErrShotAlreadyExists = errors.New("shot already exists")
	ErrCannotShootSelf   = errors.New("cannot shoot yourself")
	ErrCupidNotFound     = errors.New("cupid nomination not found")
	ErrNotInCupid        = errors.New("user is not part of this cupid nomination")
)

type SocialService struct {
	cfg             *config.Config
	shotRepo        *memory.ShotRepository
	cupidRepo       *memory.CupidRepository
	userRepo        *memory.UserRepository
	dropRepo        *memory.DropRepository
	moderationRepo  *memory.ModerationRepository
	matchingService *MatchingService
	notifService    *NotificationService
}

func NewSocialService(
	cfg *config.Config,
	shotRepo *memory.ShotRepository,
	cupidRepo *memory.CupidRepository,
	userRepo *memory.UserRepository,
	dropRepo *memory.DropRepository,
	moderationRepo *memory.ModerationRepository,
	matchingService *MatchingService,
	notifService *NotificationService,
) *SocialService {
	return &SocialService{
		cfg:             cfg,
		shotRepo:        shotRepo,
		cupidRepo:       cupidRepo,
		userRepo:        userRepo,
		dropRepo:        dropRepo,
		moderationRepo:  moderationRepo,
		matchingService: matchingService,
		notifService:    notifService,
	}
}

// ShootYourShot creates a silent one-way shot. If mutual, notifies both.
func (s *SocialService) ShootYourShot(ctx context.Context, shooterID, targetID string) (*entities.Shot, bool, error) {
	if shooterID == targetID {
		return nil, false, ErrCannotShootSelf
	}

	// Check not already shot
	existing, _ := s.shotRepo.GetByShooterAndTarget(ctx, shooterID, targetID)
	if existing != nil {
		return nil, false, ErrShotAlreadyExists
	}

	shot := &entities.Shot{
		ID:        utils.GenerateID(),
		ShooterID: shooterID,
		TargetID:  targetID,
		CreatedAt: time.Now(),
	}

	if err := s.shotRepo.Create(ctx, shot); err != nil {
		return nil, false, err
	}

	// Check for reverse shot (mutual)
	reverse, _ := s.shotRepo.GetByShooterAndTarget(ctx, targetID, shooterID)
	if reverse != nil {
		// Mutual! Create a shot drop
		score, explanation := s.matchingService.ComputeCompatibility(ctx, shooterID, targetID)
		drop := entities.NewDrop(
			utils.GenerateID(),
			shooterID,
			targetID,
			entities.DropTypeShot,
			score,
			explanation,
			s.cfg.Drop.ExpirationHours,
		)
		drop.TransitionTo(entities.DropStatusRevealed)
		drop.User1Accepted = true
		drop.User2Accepted = true
		drop.TransitionTo(entities.DropStatusPendingMutual)
		drop.TransitionTo(entities.DropStatusMatched)
		s.dropRepo.Create(ctx, drop)

		s.notifService.NotifyMutualShot(shooterID, targetID, drop.ID)
		return shot, true, nil
	}

	return shot, false, nil
}

func (s *SocialService) GetMutualShots(ctx context.Context, userID string) ([]*entities.Shot, error) {
	return s.shotRepo.GetMutualShots(ctx, userID)
}

// BrowseUsers returns paginated active quiz-complete users excluding self and blocked.
func (s *SocialService) BrowseUsers(ctx context.Context, userID string, page, pageSize int) ([]*entities.User, error) {
	blockedIDs, _ := s.moderationRepo.GetBlockedIDs(ctx, userID)
	excludeIDs := append(blockedIDs, userID)

	users, err := s.userRepo.GetActiveQuizComplete(ctx, excludeIDs)
	if err != nil {
		return nil, err
	}

	// Paginate
	start := page * pageSize
	if start >= len(users) {
		return nil, nil
	}
	end := start + pageSize
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], nil
}

type CupidNominateRequest struct {
	User1ID string `json:"user1_id"`
	User2ID string `json:"user2_id"`
}

// NominateCupid creates a cupid nomination with pre-computed compatibility.
func (s *SocialService) NominateCupid(ctx context.Context, nominatorID string, req CupidNominateRequest) (*entities.CupidNomination, error) {
	score, _ := s.matchingService.ComputeCompatibility(ctx, req.User1ID, req.User2ID)

	nom := &entities.CupidNomination{
		ID:                 utils.GenerateID(),
		NominatorID:        nominatorID,
		User1ID:            req.User1ID,
		User2ID:            req.User2ID,
		Status:             entities.CupidStatusPending,
		CompatibilityScore: score,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.cupidRepo.Create(ctx, nom); err != nil {
		return nil, err
	}

	s.notifService.NotifyCupidNomination(req.User1ID, nom.ID)
	s.notifService.NotifyCupidNomination(req.User2ID, nom.ID)

	return nom, nil
}

func (s *SocialService) AcceptCupid(ctx context.Context, userID, nomID string) (*entities.CupidNomination, error) {
	nom, err := s.cupidRepo.GetByID(ctx, nomID)
	if err != nil {
		return nil, ErrCupidNotFound
	}

	if userID != nom.User1ID && userID != nom.User2ID {
		return nil, ErrNotInCupid
	}

	nom.Accept(userID)
	if err := s.cupidRepo.Update(ctx, nom); err != nil {
		return nil, err
	}

	// If both accepted, create a cupid drop
	if nom.Status == entities.CupidStatusAccepted {
		score, explanation := s.matchingService.ComputeCompatibility(ctx, nom.User1ID, nom.User2ID)
		drop := entities.NewDrop(
			utils.GenerateID(),
			nom.User1ID,
			nom.User2ID,
			entities.DropTypeCupid,
			score,
			explanation,
			s.cfg.Drop.ExpirationHours,
		)
		drop.TransitionTo(entities.DropStatusRevealed)
		s.dropRepo.Create(ctx, drop)

		s.notifService.NotifyNewDrop(nom.User1ID, drop.ID)
		s.notifService.NotifyNewDrop(nom.User2ID, drop.ID)
	}

	return nom, nil
}

func (s *SocialService) DeclineCupid(ctx context.Context, userID, nomID string) (*entities.CupidNomination, error) {
	nom, err := s.cupidRepo.GetByID(ctx, nomID)
	if err != nil {
		return nil, ErrCupidNotFound
	}

	if userID != nom.User1ID && userID != nom.User2ID {
		return nil, ErrNotInCupid
	}

	nom.Decline(userID)
	if err := s.cupidRepo.Update(ctx, nom); err != nil {
		return nil, err
	}

	return nom, nil
}
