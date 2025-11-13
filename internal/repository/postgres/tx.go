package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"reviewer-service/internal/repository"
)

type TxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) WithTx(ctx context.Context, fn func(repository.Tx) error) error {
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txRepo := &txRepository{
		tx:        tx,
		teamRepo:  NewTeamRepository(tx),
		userRepo:  NewUserRepository(tx),
		prRepo:    NewPullRequestRepository(tx),
		statsRepo: NewStatsRepository(tx),
	}

	if err := fn(txRepo); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

type txRepository struct {
	tx        *sql.Tx
	teamRepo  repository.TeamRepository
	userRepo  repository.UserRepository
	prRepo    repository.PullRequestRepository
	statsRepo repository.StatsRepository
}

func (t *txRepository) Teams() repository.TeamRepository {
	return t.teamRepo
}

func (t *txRepository) Users() repository.UserRepository {
	return t.userRepo
}

func (t *txRepository) PullRequests() repository.PullRequestRepository {
	return t.prRepo
}

func (t *txRepository) Stats() repository.StatsRepository {
	return t.statsRepo
}

func (t *txRepository) Commit() error {
	return t.tx.Commit()
}

func (t *txRepository) Rollback() error {
	return t.tx.Rollback()
}

type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}
