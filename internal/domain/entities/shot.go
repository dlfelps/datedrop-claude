package entities

import "time"

type Shot struct {
	ID        string    `json:"id"`
	ShooterID string    `json:"shooter_id"`
	TargetID  string    `json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
}
