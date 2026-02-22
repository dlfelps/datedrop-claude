package repository

import (
	"context"
	"datedrop/internal/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	GetActiveQuizComplete(ctx context.Context, excludeIDs []string) ([]*entities.User, error)
	GetAll(ctx context.Context) ([]*entities.User, error)
}

type QuestionRepository interface {
	GetAll(ctx context.Context) ([]*entities.Question, error)
	GetByID(ctx context.Context, id string) (*entities.Question, error)
	GetByDomain(ctx context.Context, domain entities.QuestionDomain) ([]*entities.Question, error)
	Create(ctx context.Context, question *entities.Question) error
	Count(ctx context.Context) (int, error)
}

type ResponseRepository interface {
	Create(ctx context.Context, response *entities.QuizResponse) error
	GetByUserID(ctx context.Context, userID string) ([]*entities.QuizResponse, error)
	GetByUserAndQuestion(ctx context.Context, userID, questionID string) (*entities.QuizResponse, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
}

type DropRepository interface {
	Create(ctx context.Context, drop *entities.Drop) error
	GetByID(ctx context.Context, id string) (*entities.Drop, error)
	Update(ctx context.Context, drop *entities.Drop) error
	GetCurrentByUserID(ctx context.Context, userID string) (*entities.Drop, error)
	GetByUserID(ctx context.Context, userID string) ([]*entities.Drop, error)
	GetMatchedPairs(ctx context.Context, withinWeeks int) ([][2]string, error)
}

type ShotRepository interface {
	Create(ctx context.Context, shot *entities.Shot) error
	GetByShooterAndTarget(ctx context.Context, shooterID, targetID string) (*entities.Shot, error)
	GetMutualShots(ctx context.Context, userID string) ([]*entities.Shot, error)
}

type CupidRepository interface {
	Create(ctx context.Context, nom *entities.CupidNomination) error
	GetByID(ctx context.Context, id string) (*entities.CupidNomination, error)
	Update(ctx context.Context, nom *entities.CupidNomination) error
	GetByUserID(ctx context.Context, userID string) ([]*entities.CupidNomination, error)
}

type ModerationRepository interface {
	CreateBlock(ctx context.Context, block *entities.Block) error
	RemoveBlock(ctx context.Context, blockerID, blockedID string) error
	IsBlocked(ctx context.Context, userID1, userID2 string) (bool, error)
	GetBlockedIDs(ctx context.Context, userID string) ([]string, error)
	CreateReport(ctx context.Context, report *entities.Report) error
	GetReports(ctx context.Context) ([]*entities.Report, error)
}
