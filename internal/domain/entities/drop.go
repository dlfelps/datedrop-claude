package entities

import (
	"errors"
	"time"
)

type DropStatus string

const (
	DropStatusPending       DropStatus = "pending"
	DropStatusRevealed      DropStatus = "revealed"
	DropStatusPendingMutual DropStatus = "pending_mutual"
	DropStatusMatched       DropStatus = "matched"
	DropStatusCompleted     DropStatus = "completed"
	DropStatusArchived      DropStatus = "archived"
	DropStatusExpired       DropStatus = "expired"
	DropStatusDeclined      DropStatus = "declined"
)

type DropType string

const (
	DropTypeWeekly DropType = "weekly"
	DropTypeCupid  DropType = "cupid"
	DropTypeShot   DropType = "shot"
)

var dropTransitions = map[DropStatus][]DropStatus{
	DropStatusPending:       {DropStatusRevealed, DropStatusExpired, DropStatusDeclined},
	DropStatusRevealed:      {DropStatusPendingMutual, DropStatusExpired, DropStatusDeclined},
	DropStatusPendingMutual: {DropStatusMatched, DropStatusExpired, DropStatusDeclined},
	DropStatusMatched:       {DropStatusCompleted, DropStatusArchived},
	DropStatusCompleted:     {},
	DropStatusArchived:      {},
	DropStatusExpired:       {},
	DropStatusDeclined:      {},
}

type Drop struct {
	ID                string     `json:"id"`
	User1ID           string     `json:"user1_id"`
	User2ID           string     `json:"user2_id"`
	Status            DropStatus `json:"status"`
	Type              DropType   `json:"type"`
	CompatibilityScore float64   `json:"compatibility_score"`
	MatchExplanation  []string   `json:"match_explanation,omitempty"`
	User1Accepted     bool       `json:"user1_accepted"`
	User2Accepted     bool       `json:"user2_accepted"`
	ExpiresAt         time.Time  `json:"expires_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	RevealedAt        time.Time  `json:"revealed_at,omitempty"`
	MatchedAt         time.Time  `json:"matched_at,omitempty"`
}

func NewDrop(id, user1ID, user2ID string, dropType DropType, score float64, explanation []string, expirationHours int) *Drop {
	now := time.Now()
	return &Drop{
		ID:                 id,
		User1ID:            user1ID,
		User2ID:            user2ID,
		Status:             DropStatusPending,
		Type:               dropType,
		CompatibilityScore: score,
		MatchExplanation:   explanation,
		ExpiresAt:          now.Add(time.Duration(expirationHours) * time.Hour),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func (d *Drop) CanTransitionTo(newStatus DropStatus) bool {
	allowed, exists := dropTransitions[d.Status]
	if !exists {
		return false
	}
	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

func (d *Drop) TransitionTo(newStatus DropStatus) error {
	if !d.CanTransitionTo(newStatus) {
		return errors.New("invalid drop transition from " + string(d.Status) + " to " + string(newStatus))
	}
	d.Status = newStatus
	d.UpdatedAt = time.Now()

	switch newStatus {
	case DropStatusRevealed:
		d.RevealedAt = time.Now()
	case DropStatusMatched:
		d.MatchedAt = time.Now()
	}
	return nil
}

func (d *Drop) IsExpired() bool {
	return time.Now().After(d.ExpiresAt)
}

func (d *Drop) Accept(userID string) error {
	if userID == d.User1ID {
		d.User1Accepted = true
	} else if userID == d.User2ID {
		d.User2Accepted = true
	} else {
		return errors.New("user not part of this drop")
	}

	// If revealed state and first accept -> pending_mutual
	if d.Status == DropStatusRevealed {
		return d.TransitionTo(DropStatusPendingMutual)
	}
	// If pending_mutual and both accepted -> matched
	if d.Status == DropStatusPendingMutual && d.User1Accepted && d.User2Accepted {
		return d.TransitionTo(DropStatusMatched)
	}
	d.UpdatedAt = time.Now()
	return nil
}

func (d *Drop) Decline(userID string) error {
	if userID != d.User1ID && userID != d.User2ID {
		return errors.New("user not part of this drop")
	}
	return d.TransitionTo(DropStatusDeclined)
}
