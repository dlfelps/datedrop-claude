package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"

	"datedrop/internal/config"
	"datedrop/internal/domain/entities"
	"datedrop/internal/repository/memory"
	"datedrop/pkg/utils"
)

type MatchingService struct {
	cfg            *config.Config
	userRepo       *memory.UserRepository
	questionRepo   *memory.QuestionRepository
	responseRepo   *memory.ResponseRepository
	dropRepo       *memory.DropRepository
	moderationRepo *memory.ModerationRepository
	notifService   *NotificationService
}

func NewMatchingService(
	cfg *config.Config,
	userRepo *memory.UserRepository,
	questionRepo *memory.QuestionRepository,
	responseRepo *memory.ResponseRepository,
	dropRepo *memory.DropRepository,
	moderationRepo *memory.ModerationRepository,
	notifService *NotificationService,
) *MatchingService {
	return &MatchingService{
		cfg:            cfg,
		userRepo:       userRepo,
		questionRepo:   questionRepo,
		responseRepo:   responseRepo,
		dropRepo:       dropRepo,
		moderationRepo: moderationRepo,
		notifService:   notifService,
	}
}

type matchCandidate struct {
	User1ID     string
	User2ID     string
	Score       float64
	Explanation []string
}

// RunWeeklyMatching performs greedy bipartite matching on all eligible users.
func (s *MatchingService) RunWeeklyMatching(ctx context.Context) ([]*entities.Drop, error) {
	// Get all active, quiz-complete users
	users, err := s.userRepo.GetActiveQuizComplete(ctx, nil)
	if err != nil {
		return nil, err
	}

	if len(users) < 2 {
		return nil, nil
	}

	// Get previously matched pairs (lookback window)
	previousPairs, _ := s.dropRepo.GetMatchedPairs(ctx, s.cfg.Matching.LookbackWeeks)
	previousSet := make(map[string]bool)
	for _, pair := range previousPairs {
		previousSet[pair[0]+":"+pair[1]] = true
		previousSet[pair[1]+":"+pair[0]] = true
	}

	// Score all valid pairs
	var candidates []matchCandidate
	for i := 0; i < len(users); i++ {
		for j := i + 1; j < len(users); j++ {
			u1, u2 := users[i], users[j]

			// Exclusion filters
			if !u1.IsCompatibleWith(u2) {
				continue
			}
			if previousSet[u1.ID+":"+u2.ID] {
				continue
			}
			blocked, _ := s.moderationRepo.IsBlocked(ctx, u1.ID, u2.ID)
			if blocked {
				continue
			}

			score, explanation := s.computeCompatibility(ctx, u1.ID, u2.ID)
			candidates = append(candidates, matchCandidate{
				User1ID:     u1.ID,
				User2ID:     u2.ID,
				Score:       score,
				Explanation: explanation,
			})
		}
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Greedy matching
	matched := make(map[string]bool)
	var drops []*entities.Drop
	for _, c := range candidates {
		if matched[c.User1ID] || matched[c.User2ID] {
			continue
		}
		matched[c.User1ID] = true
		matched[c.User2ID] = true

		drop := entities.NewDrop(
			utils.GenerateID(),
			c.User1ID,
			c.User2ID,
			entities.DropTypeWeekly,
			c.Score,
			c.Explanation,
			s.cfg.Drop.ExpirationHours,
		)
		// Auto-reveal weekly drops
		drop.TransitionTo(entities.DropStatusRevealed)

		if err := s.dropRepo.Create(ctx, drop); err != nil {
			log.Printf("Failed to create drop: %v", err)
			continue
		}
		drops = append(drops, drop)

		s.notifService.NotifyNewDrop(c.User1ID, drop.ID)
		s.notifService.NotifyNewDrop(c.User2ID, drop.ID)
	}

	return drops, nil
}

// ComputeCompatibility calculates compatibility between two users (public for cupid).
func (s *MatchingService) ComputeCompatibility(ctx context.Context, userID1, userID2 string) (float64, []string) {
	return s.computeCompatibility(ctx, userID1, userID2)
}

func (s *MatchingService) computeCompatibility(ctx context.Context, userID1, userID2 string) (float64, []string) {
	responses1, _ := s.responseRepo.GetByUserID(ctx, userID1)
	responses2, _ := s.responseRepo.GetByUserID(ctx, userID2)

	if len(responses1) == 0 || len(responses2) == 0 {
		return 0, nil
	}

	// Index responses by question ID
	r1Map := make(map[string]*entities.QuizResponse)
	for _, r := range responses1 {
		r1Map[r.QuestionID] = r
	}
	r2Map := make(map[string]*entities.QuizResponse)
	for _, r := range responses2 {
		r2Map[r.QuestionID] = r
	}

	// Get questions for domain grouping
	questions, _ := s.questionRepo.GetAll(ctx)
	qMap := make(map[string]*entities.Question)
	for _, q := range questions {
		qMap[q.ID] = q
	}

	// Score per domain
	domainScores := map[entities.QuestionDomain]struct {
		total float64
		count int
	}{
		entities.DomainLifestyle: {},
		entities.DomainValues:    {},
		entities.DomainPolitics:  {},
	}

	type traitScore struct {
		question string
		score    float64
	}
	var topTraits []traitScore

	for qID, r1 := range r1Map {
		r2, exists := r2Map[qID]
		if !exists {
			continue
		}
		q := qMap[qID]
		if q == nil {
			continue
		}

		var rawScore float64
		switch q.ResponseType {
		case entities.ResponseTypeScale5:
			if r1.ScaleValue != nil && r2.ScaleValue != nil {
				rawScore = utils.ScaleAlignment(*r1.ScaleValue, *r2.ScaleValue, 5)
			}
		case entities.ResponseTypeScale7:
			if r1.ScaleValue != nil && r2.ScaleValue != nil {
				rawScore = utils.ScaleAlignment(*r1.ScaleValue, *r2.ScaleValue, 7)
			}
		case entities.ResponseTypeMultipleChoice:
			rawScore = utils.ExactMatchScore(r1.ChoiceValue, r2.ChoiceValue)
		case entities.ResponseTypeBoolean:
			if r1.BooleanValue != nil && r2.BooleanValue != nil {
				rawScore = utils.BooleanMatchScore(*r1.BooleanValue, *r2.BooleanValue)
			}
		}

		// Apply importance weighting (average of both users' importance)
		avgImportance := (r1.ImportanceScore + r2.ImportanceScore) / 2
		weightedScore := utils.WeightedScore(rawScore, avgImportance)

		ds := domainScores[q.Domain]
		ds.total += weightedScore
		ds.count++
		domainScores[q.Domain] = ds

		if rawScore > 0.7 {
			topTraits = append(topTraits, traitScore{question: q.Text, score: rawScore})
		}
	}

	// Calculate domain-weighted final score
	var finalScore float64
	domainWeights := map[entities.QuestionDomain]float64{
		entities.DomainLifestyle: s.cfg.Matching.LifestyleWeight,
		entities.DomainValues:    s.cfg.Matching.ValuesWeight,
		entities.DomainPolitics:  s.cfg.Matching.PoliticsWeight,
	}

	for domain, ds := range domainScores {
		if ds.count > 0 {
			avgDomainScore := ds.total / float64(ds.count)
			finalScore += avgDomainScore * domainWeights[domain]
		}
	}

	// Clamp to 0-1
	finalScore = math.Max(0, math.Min(1, finalScore))

	// Build explanation from top traits
	sort.Slice(topTraits, func(i, j int) bool {
		return topTraits[i].score > topTraits[j].score
	})
	var explanation []string
	limit := 5
	if len(topTraits) < limit {
		limit = len(topTraits)
	}
	for i := 0; i < limit; i++ {
		explanation = append(explanation, fmt.Sprintf("Both align on: %s", topTraits[i].question))
	}

	return finalScore, explanation
}
