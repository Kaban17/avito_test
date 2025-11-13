package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"

	"github.com/lib/pq"
)

type TeamRepository struct {
	db Querier
}

func NewTeamRepository(db Querier) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *entity.Team) error {
	// 1. Проверяем существование
	exists, err := r.Exists(ctx, team.Name)
	if err != nil {
		return err
	}
	if exists {
		return repository.ErrTeamExists
	}

	query := `
        INSERT INTO teams (team_name, created_at)
        VALUES ($1, NOW())
    `

	_, err = r.db.ExecContext(ctx, query, team.Name)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique violation
				return repository.ErrTeamExists
			}
		}
		return fmt.Errorf("insert team: %w", err)
	}

	if len(team.Members) > 0 {
		userRepo := NewUserRepository(r.db)
		for _, member := range team.Members {
			member.TeamName = team.Name
			err := userRepo.Upsert(ctx, member)
			if err != nil {
				return fmt.Errorf("upsert user %s: %w", member.UserID, err)
			}
		}
	}

	return nil
}

func (r *TeamRepository) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	query := `
        SELECT
            t.team_name,
            t.created_at,
            COALESCE(
                json_agg(
                    json_build_object(
                        'user_id', u.user_id,
                        'username', u.username,
                        'is_active', u.is_active
                    ) ORDER BY u.username
                ) FILTER (WHERE u.user_id IS NOT NULL),
                '[]'
            ) as members
        FROM teams t
        LEFT JOIN users u ON u.team_name = t.team_name
        WHERE t.team_name = $1
        GROUP BY t.team_name, t.created_at
    `

	var team entity.Team
	var membersJSON []byte

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&team.Name,
		&team.CreatedAt,
		&membersJSON,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query team: %w", err)
	}

	if err := json.Unmarshal(membersJSON, &team.Members); err != nil {
		return nil, fmt.Errorf("unmarshal members: %w", err)
	}

	return &team, nil
}

func (r *TeamRepository) Exists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check team exists: %w", err)
	}

	return exists, nil
}
