package entities

import "time"

type CupidStatus string

const (
	CupidStatusPending     CupidStatus = "pending"
	CupidStatusUser1Accept CupidStatus = "user1_accepted"
	CupidStatusUser2Accept CupidStatus = "user2_accepted"
	CupidStatusAccepted    CupidStatus = "accepted"
	CupidStatusDeclined    CupidStatus = "declined"
)

type CupidNomination struct {
	ID                 string      `json:"id"`
	NominatorID        string      `json:"nominator_id"`
	User1ID            string      `json:"user1_id"`
	User2ID            string      `json:"user2_id"`
	Status             CupidStatus `json:"status"`
	CompatibilityScore float64     `json:"compatibility_score"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

func (cn *CupidNomination) Accept(userID string) {
	if userID == cn.User1ID {
		if cn.Status == CupidStatusPending {
			cn.Status = CupidStatusUser1Accept
		} else if cn.Status == CupidStatusUser2Accept {
			cn.Status = CupidStatusAccepted
		}
	} else if userID == cn.User2ID {
		if cn.Status == CupidStatusPending {
			cn.Status = CupidStatusUser2Accept
		} else if cn.Status == CupidStatusUser1Accept {
			cn.Status = CupidStatusAccepted
		}
	}
	cn.UpdatedAt = time.Now()
}

func (cn *CupidNomination) Decline(userID string) {
	if userID == cn.User1ID || userID == cn.User2ID {
		cn.Status = CupidStatusDeclined
		cn.UpdatedAt = time.Now()
	}
}
