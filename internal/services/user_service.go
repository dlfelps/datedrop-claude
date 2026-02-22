package services

import (
	"context"
	"errors"
	"time"

	"datedrop/internal/domain/entities"
	"datedrop/internal/repository/memory"
	"datedrop/pkg/utils"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrEmailTaken     = errors.New("email already registered")
	ErrInvalidEmail   = errors.New("must use a .edu email address")
	ErrTooYoung       = errors.New("must be at least 18 years old")
	ErrNotAuthorized  = errors.New("not authorized")
)

type UserService struct {
	userRepo *memory.UserRepository
}

func NewUserService(userRepo *memory.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type CreateUserRequest struct {
	Email        string               `json:"email"`
	Name         string               `json:"name"`
	DateOfBirth  time.Time            `json:"date_of_birth"`
	Gender       entities.Gender      `json:"gender"`
	Orientations []entities.Orientation `json:"orientations"`
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*entities.User, error) {
	if !utils.IsEduEmail(req.Email) {
		return nil, ErrInvalidEmail
	}
	if !utils.IsAtLeast18(req.DateOfBirth) {
		return nil, ErrTooYoung
	}

	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	user := entities.NewUser(
		utils.GenerateID(),
		req.Email,
		req.Name,
		req.DateOfBirth,
		req.Gender,
		req.Orientations,
	)

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login returns the user for the given email (mock auth — no password).
func (s *UserService) Login(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

type UpdateUserRequest struct {
	Name   *string   `json:"name,omitempty"`
	Bio    *string   `json:"bio,omitempty"`
	Photos []string  `json:"photos,omitempty"`
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.Photos != nil {
		user.Photos = req.Photos
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
