package utils

import "math"

// ScaleAlignment returns a 0-1 score for how close two scale responses are.
// maxScale is the maximum value on the scale (e.g., 5 or 7).
func ScaleAlignment(a, b, maxScale int) float64 {
	diff := math.Abs(float64(a - b))
	return 1.0 - (diff / float64(maxScale-1))
}

// ExactMatchScore returns 1.0 if values match, 0.0 otherwise.
func ExactMatchScore(a, b string) float64 {
	if a == b {
		return 1.0
	}
	return 0.0
}

// BooleanMatchScore returns 1.0 if booleans match, 0.0 otherwise.
func BooleanMatchScore(a, b bool) float64 {
	if a == b {
		return 1.0
	}
	return 0.0
}

// WeightedScore applies importance weighting to a raw alignment score.
// When importance is high and alignment is low, the penalty is amplified.
// importance is 1-5, rawScore is 0-1.
func WeightedScore(rawScore float64, importance int) float64 {
	weight := float64(importance) / 5.0
	if rawScore < 0.5 {
		// Mismatch on high-importance = bigger penalty
		return rawScore * (1.0 - weight*0.5)
	}
	return rawScore * (0.5 + weight*0.5)
}
