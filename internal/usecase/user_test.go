package usecase

import (
	"context"
	"testing"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"
)

// Mock implementation of TxManager for testing
type mockTxManager struct {
	withTxFn func(context.Context, func(repository.Tx) error) error
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(repository.Tx) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(&mockTx{})
}

// Mock implementation of Tx interface for testing
type mockTx struct {
	usersRepo repository.UserRepository
	prRepo    repository.PullRequestRepository
	statsRepo repository.StatsRepository
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
	getByIDFn   func(context.Context, string) (*entity.User, error)
	setActiveFn func(context.Context, string, bool) error
	getByTeamFn func(context.Context, string) ([]*entity.User, error)
	getActiveFn func(context.Context, string, string) ([]*entity.User, error)
}

func (m *mockUsersRepo) Create(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUsersRepo) Update(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUsersRepo) Upsert(ctx context.Context, user *entity.User) error {
	return nil
}

func (m *mockUsersRepo) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID)
	}
	return &entity.User{UserID: userID, IsActive: true}, nil
}

func (m *mockUsersRepo) SetActive(ctx context.Context, userID string, isActive bool) error {
	if m.setActiveFn != nil {
		return m.setActiveFn(ctx, userID, isActive)
	}
	return nil
}

func (m *mockUsersRepo) GetByTeam(ctx context.Context, teamName string) ([]*entity.User, error) {
	if m.getByTeamFn != nil {
		return m.getByTeamFn(ctx, teamName)
	}
	return []*entity.User{}, nil
}

func (m *mockUsersRepo) GetActiveByTeam(ctx context.Context, teamName, excludeID string) ([]*entity.User, error) {
	if m.getActiveFn != nil {
		return m.getActiveFn(ctx, teamName, excludeID)
	}
	return []*entity.User{}, nil
}

func (m *mockUsersRepo) BulkDeactivate(ctx context.Context, userIDs []string) error {
	return nil
}

type mockPRRepo struct {
	getByReviewerFn func(context.Context, string) ([]*entity.PullRequest, error)
	getByIDFn       func(context.Context, string) (*entity.PullRequest, error)
}

func (m *mockPRRepo) Create(ctx context.Context, pr *entity.PullRequest) error {
	return nil
}

func (m *mockPRRepo) Update(ctx context.Context, pr *entity.PullRequest) error {
	return nil
}

func (m *mockPRRepo) GetByID(ctx context.Context, id string) (*entity.PullRequest, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockPRRepo) GetByIDForUpdate(ctx context.Context, id string) (*entity.PullRequest, error) {
	return nil, nil
}

func (m *mockPRRepo) GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	if m.getByReviewerFn != nil {
		return m.getByReviewerFn(ctx, userID)
	}
	return []*entity.PullRequest{}, nil
}

func (m *mockPRRepo) GetOpenByReviewers(ctx context.Context, userIDs []string) ([]*entity.PullRequest, error) {
	return []*entity.PullRequest{}, nil
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

func TestUserUseCase_SetActive(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expectedUser := &entity.User{UserID: "user123", IsActive: true}
		txManager := &mockTxManager{
			withTxFn: func(ctx context.Context, fn func(repository.Tx) error) error {
				return fn(&mockTx{
					usersRepo: &mockUsersRepo{
						setActiveFn: func(ctx context.Context, userID string, isActive bool) error {
							return nil
						},
						getByIDFn: func(ctx context.Context, userID string) (*entity.User, error) {
							return expectedUser, nil
						},
					},
				})
			},
		}

		usecase := NewUserUseCase(txManager)

		user, err := usecase.SetActive(ctx, "user123", true)
		if err != nil {
			t.Fatalf("SetActive() error = %v", err)
		}

		if user != expectedUser {
			t.Errorf("SetActive() = %v, want %v", user, expectedUser)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		txManager := &mockTxManager{
			withTxFn: func(ctx context.Context, fn func(repository.Tx) error) error {
				return fn(&mockTx{
					usersRepo: &mockUsersRepo{
						setActiveFn: func(ctx context.Context, userID string, isActive bool) error {
							return entity.ErrUserNotFound
						},
						getByIDFn: func(ctx context.Context, userID string) (*entity.User, error) {
							return nil, entity.ErrUserNotFound
						},
					},
				})
			},
		}

		usecase := NewUserUseCase(txManager)

		_, err := usecase.SetActive(ctx, "user123", true)
		if err == nil {
			t.Error("SetActive() expected error, got nil")
		}
	})
}

func TestUserUseCase_GetReviews(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expectedPRs := []*entity.PullRequest{
			{ID: "pr1", AssignedReviewers: []string{"user123"}},
			{ID: "pr2", AssignedReviewers: []string{"user123"}},
		}

		txManager := &mockTxManager{
			withTxFn: func(ctx context.Context, fn func(repository.Tx) error) error {
				return fn(&mockTx{
					usersRepo: &mockUsersRepo{
						getByIDFn: func(ctx context.Context, userID string) (*entity.User, error) {
							return &entity.User{UserID: userID}, nil
						},
					},
					prRepo: &mockPRRepo{
						getByReviewerFn: func(ctx context.Context, reviewerID string) ([]*entity.PullRequest, error) {
							return expectedPRs, nil
						},
					},
				})
			},
		}

		usecase := NewUserUseCase(txManager)

		prs, err := usecase.GetReviews(ctx, "user123")
		if err != nil {
			t.Fatalf("GetReviews() error = %v", err)
		}

		if len(prs) != len(expectedPRs) {
			t.Errorf("GetReviews() = %v PRs, want %v", len(prs), len(expectedPRs))
		}
	})

	t.Run("User not found", func(t *testing.T) {
		txManager := &mockTxManager{
			withTxFn: func(ctx context.Context, fn func(repository.Tx) error) error {
				return fn(&mockTx{
					usersRepo: &mockUsersRepo{
						getByIDFn: func(ctx context.Context, userID string) (*entity.User, error) {
							return nil, entity.ErrUserNotFound
						},
					},
				})
			},
		}

		usecase := NewUserUseCase(txManager)

		_, err := usecase.GetReviews(ctx, "user123")
		if err == nil {
			t.Error("GetReviews() expected error, got nil")
		}
	})
}
