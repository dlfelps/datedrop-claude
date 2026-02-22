package entities

import "time"

type Block struct {
	ID        string    `json:"id"`
	BlockerID string    `json:"blocker_id"`
	BlockedID string    `json:"blocked_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ReportCategory string

const (
	ReportCategoryHarassment    ReportCategory = "harassment"
	ReportCategoryInappropriate ReportCategory = "inappropriate"
	ReportCategorySpam          ReportCategory = "spam"
	ReportCategoryFakeProfile   ReportCategory = "fake_profile"
	ReportCategoryOther         ReportCategory = "other"
)

type ReportStatus string

const (
	ReportStatusOpen     ReportStatus = "open"
	ReportStatusReviewed ReportStatus = "reviewed"
	ReportStatusResolved ReportStatus = "resolved"
)

type Report struct {
	ID         string         `json:"id"`
	ReporterID string         `json:"reporter_id"`
	ReportedID string         `json:"reported_id"`
	Category   ReportCategory `json:"category"`
	Details    string         `json:"details"`
	Status     ReportStatus   `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
}
