package entities

import "time"

type QuizResponse struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	QuestionID      string `json:"question_id"`
	ScaleValue      *int   `json:"scale_value,omitempty"`
	ChoiceValue     string `json:"choice_value,omitempty"`
	BooleanValue    *bool  `json:"boolean_value,omitempty"`
	ImportanceScore int    `json:"importance_score"` // 1-5
	CreatedAt       time.Time `json:"created_at"`
}
