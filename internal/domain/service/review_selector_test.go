package service

import (
	"context"
	"testing"
	"time"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"
)

// Mock implementation of Tx interface for testing
type mockTx struct {
	usersRepo repository.UserRepository
	statsRepo repository.StatsRepository
	prRepo    repository.PullRequestRepository
	teamRepo  repository.TeamRepository
}

func (m *mockTx) Teams() repository.TeamRepository {
	return m.teamRepo
}

func (m *mockTx) Users() repository.UserRepository {
	return m.usersRepo
}

func (m *mockTx) Stats() repository.StatsRepository {
	return m.statsRepo
}

func (m *mockTx) PullRequests() repository.PullRequestRepository {
	return m.prRepo
}

func (m *mockTx) Commit() error {
	return nil
}

func (m *mockTx) Rollback() error {
	return nil
}

// Mock implementations of repository interfaces
type mockUsersRepo struct {
	users map[string]*entity.User
}

func (m *mockUsersRepo) Create(ctx context.Context, user *entity.User) error {
	m.users[user.UserID] = user
	return nil
}

func (m *mockUsersRepo) Update(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUsersRepo) Upsert(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUsersRepo) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	user, exists := m.users[userID]
	if !exists {
		return nil, entity.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUsersRepo) GetByTeam(ctx context.Context, teamName string) ([]*entity.User, error) {
	var users []*entity.User
	for _, user := range m.users {
		if user.TeamName == teamName {
			users = append(users, user)
		}
	}
	return users, nil
}

func (m *mockUsersRepo) GetActiveByTeam(ctx context.Context, teamName, excludeID string) ([]*entity.User, error) {
	var users []*entity.User
	for _, user := range m.users {
		if user.TeamName == teamName && user.IsActive && user.UserID != excludeID {
			users = append(users, user)
		}
	}
	return users, nil
}

func (m *mockUsersRepo) SetActive(ctx context.Context, userID string, isActive bool) error {
	user, exists := m.users[userID]
	if !exists {
		return entity.ErrUserNotFound
	}
	user.IsActive = isActive
	user.UpdatedAt = time.Now()
	return nil
}

func (m *mockUsersRepo) BulkDeactivate(ctx context.Context, userIDs []string) error {
	return nil
}

type mockStatsRepo struct {
	workload map[string]int
}

func (m *mockStatsRepo) GetWorkload(ctx context.Context, userIDs []string) (map[string]int, error) {
	result := make(map[string]int)
	for _, id := range userIDs {
		if load, exists := m.workload[id]; exists {
			result[id] = load
		}
	}
	return result, nil
}

func (m *mockStatsRepo) IncrementAssignment(ctx context.Context, userID string) error {
	return nil
}

func (m *mockStatsRepo) GetUserStats(ctx context.Context, userID string) (*entity.UserStats, error) {
	return nil, nil
}

func (m *mockStatsRepo) GetTeamStats(ctx context.Context) ([]*entity.TeamStats, error) {
	return nil, nil
}

type mockPRRepo struct{}

func (m *mockPRRepo) Create(ctx context.Context, pr *entity.PullRequest) error {
	return nil
}

func (m *mockPRRepo) Update(ctx context.Context, pr *entity.PullRequest) error {
	return nil
}

func (m *mockPRRepo) GetByID(ctx context.Context, id string) (*entity.PullRequest, error) {
	return nil, nil
}

func (m *mockPRRepo) GetByIDForUpdate(ctx context.Context, id string) (*entity.PullRequest, error) {
	return nil, nil
}

func (m *mockPRRepo) GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	return nil, nil
}

func (m *mockPRRepo) GetOpenByReviewers(ctx context.Context, userIDs []string) ([]*entity.PullRequest, error) {
	return nil, nil
}

func (m *mockPRRepo) AssignReviewers(ctx context.Context, prID string, userIDs []string) error {
	return nil
}

func (m *mockPRRepo) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	return nil
}

func (m *mockPRRepo) IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
	return false, nil
}

type mockTeamRepo struct{}

func (m *mockTeamRepo) Create(ctx context.Context, team *entity.Team) error {
	return nil
}

func (m *mockTeamRepo) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	return nil, nil
}

func (m *mockTeamRepo) Exists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

// Mock implementation of Txable interface for testing
type mockTxManager struct {
	withTxFn func(context.Context, func(repository.Tx) error) error
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(repository.Tx) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}

	// Default implementation
	return fn(&mockTx{
		usersRepo: &mockUsersRepo{
			users: map[string]*entity.User{
				"user1": {UserID: "user1", Username: "alice", TeamName: "backend", IsActive: true},
				"user2": {UserID: "user2", Username: "bob", TeamName: "backend", IsActive: true},
				"user3": {UserID: "user3", Username: "charlie", TeamName: "backend", IsActive: true},
				"user4": {UserID: "user4", Username: "dave", TeamName: "backend", IsActive: false},
			},
		},
		statsRepo: &mockStatsRepo{
			workload: map[string]int{
				"user1": 5,
				"user2": 3,
				"user3": 7,
			},
		},
	})
}

func TestReviewerSelector_Select(t *testing.T) {
	selector := NewReviewerSelector()
	txManager := &mockTxManager{}

	ctx := context.Background()

	// Тест 1: Нормальный случай - выбираем 2 ревьюеров
	t.Run("Select reviewers", func(t *testing.T) {
		var selected []*entity.User
		err := txManager.WithTx(ctx, func(tx repository.Tx) error {
			var err error
			selected, err = selector.Select(ctx, tx, "backend", "user5")
			return err
		})

		if err != nil {
			t.Fatalf("Select() error = %v", err)
		}

		// Должно выбрать 2 ревьюера
		if len(selected) != 2 {
			t.Errorf("Select() = %v reviewers, want 2", len(selected))
		}

		// Проверяем, что не выбрали автора
		for _, user := range selected {
			if user.UserID == "user5" {
				t.Error("Select() should not select author as reviewer")
			}
		}
	})

	// Тест 2: Нет активных кандидатов
	t.Run("No active candidates", func(t *testing.T) {
		var selected []*entity.User
		err := txManager.WithTx(ctx, func(tx repository.Tx) error {
			var err error
			// Все пользователи в команде неактивны или это автор
			selected, err = selector.Select(ctx, tx, "frontend", "user5")
			return err
		})

		if err != nil {
			t.Fatalf("Select() error = %v", err)
		}

		// Должно выбрать 0 ревьюеров
		if len(selected) != 0 {
			t.Errorf("Select() = %v reviewers, want 0", len(selected))
		}
	})
}

func TestReviewerSelector_SelectReplacement(t *testing.T) {
	selector := NewReviewerSelector()
	txManager := &mockTxManager{}

	ctx := context.Background()

	// Тест 1: Нормальный случай - выбираем замену
	t.Run("Select replacement", func(t *testing.T) {
		var selected *entity.User
		err := txManager.WithTx(ctx, func(tx repository.Tx) error {
			var err error
			selected, err = selector.SelectReplacement(ctx, tx, "backend", []string{"user1"})
			return err
		})

		if err != nil {
			t.Fatalf("SelectReplacement() error = %v", err)
		}

		if selected == nil {
			t.Fatal("SelectReplacement() = nil, want user")
		}

		// Проверяем, что выбрали пользователя с минимальной нагрузкой
		// user2 имеет нагрузку 3, что меньше чем у user1 (5) и user3 (7)
		if selected.UserID != "user2" {
			t.Errorf("SelectReplacement() = %v, want user2 (user with least workload)", selected.UserID)
		}

		// Проверяем, что не выбрали исключенного пользователя
		if selected.UserID == "user1" {
			t.Error("SelectReplacement() should not select excluded user")
		}
	})

	// Тест 2: Нет доступных кандидатов
	t.Run("No candidates", func(t *testing.T) {
		var selected *entity.User
		err := txManager.WithTx(ctx, func(tx repository.Tx) error {
			var err error
			// Исключаем всех активных пользователей
			selected, err = selector.SelectReplacement(ctx, tx, "backend", []string{"user1", "user2", "user3"})
			return err
		})

		if err != nil {
			t.Fatalf("SelectReplacement() error = %v", err)
		}

		// Должно вернуть nil
		if selected != nil {
			t.Errorf("SelectReplacement() = %v, want nil", selected)
		}
	})
}
