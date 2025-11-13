package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"

	"github.com/lib/pq"
)

type StatsRepository struct {
	db Querier
}

func NewStatsRepository(db Querier) *StatsRepository {
	return &StatsRepository{db: db}
}

// GetWorkload возвращает количество открытых PR для каждого пользователя
// Это ключевой метод для fair distribution
func (r *StatsRepository) GetWorkload(ctx context.Context, userIDs []string) (map[string]int, error) {
	if len(userIDs) == 0 {
		return make(map[string]int), nil
	}

	query := `
        SELECT
            r.user_id,
            COUNT(*) as open_count
        FROM pr_reviewers r
        INNER JOIN pull_requests pr ON r.pull_request_id = pr.pull_request_id
        WHERE r.user_id = ANY($1) AND pr.status = 'OPEN'
        GROUP BY r.user_id
    `

	rows, err := r.db.QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, fmt.Errorf("query workload: %w", err)
	}
	defer rows.Close()

	workload := make(map[string]int)

	// Инициализируем всех нулями
	for _, userID := range userIDs {
		workload[userID] = 0
	}

	// Заполняем реальными значениями
	for rows.Next() {
		var userID string
		var count int

		if err := rows.Scan(&userID, &count); err != nil {
			return nil, fmt.Errorf("scan workload: %w", err)
		}

		workload[userID] = count
	}

	return workload, rows.Err()
}

// IncrementAssignment увеличивает счётчик назначений (для статистики)
func (r *StatsRepository) IncrementAssignment(ctx context.Context, userID string) error {
	query := `
        INSERT INTO assignment_stats (user_id, assignment_count, last_assigned_at)
        VALUES ($1, 1, NOW())
        ON CONFLICT (user_id)
        DO UPDATE SET
            assignment_count = assignment_stats.assignment_count + 1,
            last_assigned_at = NOW()
    `

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("increment assignment: %w", err)
	}

	return nil
}

// GetUserStats возвращает детальную статистику по пользователю
func (r *StatsRepository) GetUserStats(ctx context.Context, userID string) (*entity.UserStats, error) {
	query := `
        SELECT
            u.user_id,
            u.username,
            u.team_name,
            COALESCE(
                (SELECT COUNT(*)
                 FROM pr_reviewers r
                 INNER JOIN pull_requests pr ON r.pull_request_id = pr.pull_request_id
                 WHERE r.user_id = u.user_id AND pr.status = 'OPEN'),
                0
            ) as open_reviews,
            COALESCE(
                (SELECT COUNT(*)
                 FROM pr_reviewers r
                 WHERE r.user_id = u.user_id),
                0
            ) as total_reviews,
            COALESCE(s.assignment_count, 0) as assignment_count,
            s.last_assigned_at
        FROM users u
        LEFT JOIN assignment_stats s ON u.user_id = s.user_id
        WHERE u.user_id = $1
    `

	var stats entity.UserStats
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.UserID,
		&stats.Username,
		&stats.TeamName,
		&stats.OpenReviews,
		&stats.TotalReviews,
		&stats.AssignmentCount,
		&stats.LastAssignedAt,
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user stats: %w", err)
	}

	return &stats, nil
}

// GetTeamStats возвращает статистику по всем командам
func (r *StatsRepository) GetTeamStats(ctx context.Context) ([]*entity.TeamStats, error) {
	query := `
        SELECT
            t.team_name,
            COUNT(DISTINCT u.user_id) as total_members,
            COUNT(DISTINCT u.user_id) FILTER (WHERE u.is_active = true) as active_members,
            COALESCE(
                (SELECT COUNT(DISTINCT pr.pull_request_id)
                 FROM pull_requests pr
                 INNER JOIN users au ON pr.author_id = au.user_id
                 WHERE au.team_name = t.team_name),
                0
            ) as total_prs,
            COALESCE(
                (SELECT COUNT(DISTINCT pr.pull_request_id)
                 FROM pull_requests pr
                 INNER JOIN users au ON pr.author_id = au.user_id
                 WHERE au.team_name = t.team_name AND pr.status = 'OPEN'),
                0
            ) as open_prs
        FROM teams t
        LEFT JOIN users u ON t.team_name = u.team_name
        GROUP BY t.team_name
        ORDER BY t.team_name
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query team stats: %w", err)
	}
	defer rows.Close()

	var stats []*entity.TeamStats
	for rows.Next() {
		var s entity.TeamStats
		if err := rows.Scan(
			&s.TeamName,
			&s.TotalMembers,
			&s.ActiveMembers,
			&s.TotalPRs,
			&s.OpenPRs,
		); err != nil {
			return nil, fmt.Errorf("scan team stats: %w", err)
		}
		stats = append(stats, &s)
	}

	return stats, rows.Err()
}
