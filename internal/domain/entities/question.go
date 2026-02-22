package entities

type QuestionDomain string

const (
	DomainLifestyle QuestionDomain = "lifestyle"
	DomainValues    QuestionDomain = "values"
	DomainPolitics  QuestionDomain = "politics"
)

type ResponseType string

const (
	ResponseTypeScale5         ResponseType = "scale_5"
	ResponseTypeScale7         ResponseType = "scale_7"
	ResponseTypeMultipleChoice ResponseType = "multiple_choice"
	ResponseTypeBoolean        ResponseType = "boolean"
)

type Question struct {
	ID           string         `json:"id"`
	Text         string         `json:"text"`
	Domain       QuestionDomain `json:"domain"`
	ResponseType ResponseType   `json:"response_type"`
	Options      []string       `json:"options,omitempty"` // for multiple_choice
	Version      int            `json:"version"`
	OrderIndex   int            `json:"order_index"`
}
