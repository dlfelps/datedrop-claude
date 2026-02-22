package entities

import "time"

type Gender string

const (
	GenderMale      Gender = "male"
	GenderFemale    Gender = "female"
	GenderNonBinary Gender = "non_binary"
)

type Orientation string

const (
	OrientationStraight Orientation = "straight"
	OrientationGay      Orientation = "gay"
	OrientationBisexual Orientation = "bisexual"
)

type User struct {
	ID             string        `json:"id"`
	Email          string        `json:"email"`
	Name           string        `json:"name"`
	Bio            string        `json:"bio,omitempty"`
	Photos         []string      `json:"photos,omitempty"`
	DateOfBirth    time.Time     `json:"date_of_birth"`
	Gender         Gender        `json:"gender"`
	Orientations   []Orientation `json:"orientations"`
	QuizCompleted  bool          `json:"quiz_completed"`
	IsActive       bool          `json:"is_active"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

func NewUser(id, email, name string, dob time.Time, gender Gender, orientations []Orientation) *User {
	now := time.Now()
	return &User{
		ID:           id,
		Email:        email,
		Name:         name,
		DateOfBirth:  dob,
		Gender:       gender,
		Orientations: orientations,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// IsCompatibleWith checks gender/orientation compatibility between two users.
func (u *User) IsCompatibleWith(other *User) bool {
	return u.isAttractedTo(other) && other.isAttractedTo(u)
}

func (u *User) isAttractedTo(other *User) bool {
	for _, o := range u.Orientations {
		switch o {
		case OrientationStraight:
			if u.Gender == GenderMale && other.Gender == GenderFemale {
				return true
			}
			if u.Gender == GenderFemale && other.Gender == GenderMale {
				return true
			}
			if u.Gender == GenderNonBinary {
				return true
			}
		case OrientationGay:
			if u.Gender == other.Gender {
				return true
			}
			if u.Gender == GenderNonBinary || other.Gender == GenderNonBinary {
				return true
			}
		case OrientationBisexual:
			return true
		}
	}
	return false
}
