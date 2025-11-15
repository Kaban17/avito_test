package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"

	"github.com/lib/pq"
)

type UserRepository struct {
	db Querier
}

func NewUserRepository(db Querier) *UserRepository {
	return &UserRepository{db: db}
}

// Upsert создаёт или обновляет пользователя
func (r *UserRepository) Upsert(ctx context.Context, user *entity.User) error {
	query := `
        INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        ON CONFLICT (user_id)
        DO UPDATE SET
            username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active,
            updated_at = NOW()
    `

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.TeamName,
		user.IsActive,
	)

	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	return nil
}

// Create создаёт пользователя
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
        INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
    `

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.TeamName,
		user.IsActive,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique violation
				return repository.ErrUserExists
			}
			if pqErr.Code == "23503" { // foreign key violation
				return repository.ErrTeamNotFound
			}
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
        UPDATE users
        SET username = $2, team_name = $3, is_active = $4, updated_at = NOW()
        WHERE user_id = $1
    `

	result, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.TeamName,
		user.IsActive,
	)

	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return repository.ErrNotFound
	}

	return nil
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	query := `
        SELECT user_id, username, team_name, is_active, created_at, updated_at
        FROM users
        WHERE user_id = $1
    `

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByTeam(ctx context.Context, teamName string) ([]*entity.User, error) {
	query := `
        SELECT user_id, username, team_name, is_active, created_at, updated_at
        FROM users
        WHERE team_name = $1
        ORDER BY username
    `

	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("query users by team: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error or handle it appropriately
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

func (r *UserRepository) GetActiveByTeam(
	ctx context.Context,
	teamName string,
	excludeUserID string,
) ([]*entity.User, error) {
	query := `
        SELECT user_id, username, team_name, is_active, created_at, updated_at
        FROM users
        WHERE team_name = $1
          AND is_active = true
          AND user_id != $2
        ORDER BY username
    `

	rows, err := r.db.QueryContext(ctx, query, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("query active users: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error or handle it appropriately
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

func (r *UserRepository) SetActive(ctx context.Context, userID string, isActive bool) error {
	query := `
        UPDATE users
        SET is_active = $2, updated_at = NOW()
        WHERE user_id = $1
    `

	result, err := r.db.ExecContext(ctx, query, userID, isActive)
	if err != nil {
		return fmt.Errorf("set user active: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *UserRepository) BulkDeactivate(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	query := `
        UPDATE users
        SET is_active = false, updated_at = NOW()
        WHERE user_id = ANY($1)
    `

	_, err := r.db.ExecContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return fmt.Errorf("bulk deactivate users: %w", err)
	}

	return nil
}
