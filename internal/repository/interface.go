package repository

import (
	"context"
	"reviewer-service/internal/domain/entity"
)

// Txable - интерфейс для выполнения операций в транзакции
type Txable interface {
	WithTx(ctx context.Context, fn func(Tx) error) error
}

// Tx - транзакционный контекст с доступом ко всем репозиториям
type Tx interface {
	Teams() TeamRepository
	Users() UserRepository
	PullRequests() PullRequestRepository
	Stats() StatsRepository

	Commit() error
	Rollback() error
}

// TeamRepository - операции с командами
type TeamRepository interface {
	Create(ctx context.Context, team *entity.Team) error
	GetByName(ctx context.Context, name string) (*entity.Team, error)
	Exists(ctx context.Context, name string) (bool, error)
}

// UserRepository - операции с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Upsert(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, userID string) (*entity.User, error)
	GetByTeam(ctx context.Context, teamName string) ([]*entity.User, error)
	GetActiveByTeam(ctx context.Context, teamName string, excludeUserID string) ([]*entity.User, error)
	SetActive(ctx context.Context, userID string, isActive bool) error
	BulkDeactivate(ctx context.Context, userIDs []string) error
}

// PullRequestRepository - операции с PR
type PullRequestRepository interface {
	Create(ctx context.Context, pr *entity.PullRequest) error
	Update(ctx context.Context, pr *entity.PullRequest) error
	GetByID(ctx context.Context, prID string) (*entity.PullRequest, error)
	GetByIDForUpdate(ctx context.Context, prID string) (*entity.PullRequest, error)
	GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error)
	GetOpenByReviewers(ctx context.Context, userIDs []string) ([]*entity.PullRequest, error)

	// Работа с ревьюверами
	AssignReviewers(ctx context.Context, prID string, userIDs []string) error
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error)
}

// StatsRepository - статистика и аналитика
type StatsRepository interface {
	GetWorkload(ctx context.Context, userIDs []string) (map[string]int, error)
	IncrementAssignment(ctx context.Context, userID string) error
	GetUserStats(ctx context.Context, userID string) (*entity.UserStats, error)
	GetTeamStats(ctx context.Context) ([]*entity.TeamStats, error)
}
