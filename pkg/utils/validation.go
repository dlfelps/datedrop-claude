package utils

import (
	"strings"
	"time"
)

func IsEduEmail(email string) bool {
	return strings.HasSuffix(strings.ToLower(email), ".edu")
}

func IsAtLeast18(dob time.Time) bool {
	now := time.Now()
	age := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		age--
	}
	return age >= 18
}
