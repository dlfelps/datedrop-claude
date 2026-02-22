package handlers

import (
	"time"

	"datedrop/internal/domain/entities"
)

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func parseGender(s string) entities.Gender {
	switch s {
	case "male":
		return entities.GenderMale
	case "female":
		return entities.GenderFemale
	case "non_binary":
		return entities.GenderNonBinary
	default:
		return entities.Gender(s)
	}
}

func parseOrientations(ss []string) []entities.Orientation {
	result := make([]entities.Orientation, len(ss))
	for i, s := range ss {
		result[i] = entities.Orientation(s)
	}
	return result
}
