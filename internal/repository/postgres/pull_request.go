package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"

	"github.com/lib/pq"
)

type PullRequestRepository struct {
	db Querier
}

func NewPullRequestRepository(db Querier) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) Create(ctx context.Context, pr *entity.PullRequest) error {
	query := `
        INSERT INTO pull_requests (
            pull_request_id,
            pull_request_name,
            author_id,
            status,
            created_at,
            version
        )
        VALUES ($1, $2, $3, $4, NOW(), 1)
    `

	_, err := r.db.ExecContext(ctx, query,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique violation
				return repository.ErrPRExists
			}
			if pqErr.Code == "23503" { // foreign key violation
				return repository.ErrUserNotFound
			}
		}
		return fmt.Errorf("insert pr: %w", err)
	}

	return nil
}

func (r *PullRequestRepository) Update(ctx context.Context, pr *entity.PullRequest) error {
	query := `
        UPDATE pull_requests
        SET
            pull_request_name = $2,
            status = $3,
            merged_at = $4,
            version = version + 1
        WHERE pull_request_id = $1 AND version = $5
    `

	result, err := r.db.ExecContext(ctx, query,
		pr.ID,
		pr.Name,
		pr.Status,
		pr.MergedAt,
		pr.Version,
	)

	if err != nil {
		return fmt.Errorf("update pr: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return repository.ErrOptimisticLock
	}

	pr.Version++
	return nil
}

func (r *PullRequestRepository) GetByID(ctx context.Context, prID string) (*entity.PullRequest, error) {
	query := `
        SELECT
            pr.pull_request_id,
            pr.pull_request_name,
            pr.author_id,
            pr.status,
            pr.created_at,
            pr.merged_at,
            pr.version,
            COALESCE(
                array_agg(r.user_id ORDER BY r.assigned_at)
                FILTER (WHERE r.user_id IS NOT NULL),
                '{}'
            ) as reviewer_ids
        FROM pull_requests pr
        LEFT JOIN pr_reviewers r ON pr.pull_request_id = r.pull_request_id
        WHERE pr.pull_request_id = $1
        GROUP BY pr.pull_request_id
    `

	var pr entity.PullRequest
	var reviewerIDs []string

	err := r.db.QueryRowContext(ctx, query, prID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
		&pr.Version,
		pq.Array(&reviewerIDs),
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query pr: %w", err)
	}

	pr.AssignedReviewers = reviewerIDs
	return &pr, nil
}

func (r *PullRequestRepository) GetByIDForUpdate(ctx context.Context, prID string) (*entity.PullRequest, error) {
	// Сначала блокируем строку PR
	lockQuery := `
        SELECT pull_request_id
        FROM pull_requests
        WHERE pull_request_id = $1
        FOR UPDATE
    `

	err := r.db.QueryRowContext(ctx, lockQuery, prID).Scan(&prID)
	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lock pr: %w", err)
	}

	// Затем получаем полную информацию о PR
	query := `
        SELECT
            pr.pull_request_id,
            pr.pull_request_name,
            pr.author_id,
            pr.status,
            pr.created_at,
            pr.merged_at,
            pr.version,
            COALESCE(
                array_agg(r.user_id ORDER BY r.assigned_at)
                FILTER (WHERE r.user_id IS NOT NULL),
                '{}'
            ) as reviewer_ids
        FROM pull_requests pr
        LEFT JOIN pr_reviewers r ON pr.pull_request_id = r.pull_request_id
        WHERE pr.pull_request_id = $1
        GROUP BY pr.pull_request_id
    `

	var pr entity.PullRequest
	var reviewerIDs []string

	err = r.db.QueryRowContext(ctx, query, prID).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
		&pr.Version,
		pq.Array(&reviewerIDs),
	)

	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query pr: %w", err)
	}

	pr.AssignedReviewers = reviewerIDs
	return &pr, nil
}

func (r *PullRequestRepository) GetByReviewer(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	query := `
        SELECT DISTINCT
            pr.pull_request_id,
            pr.pull_request_name,
            pr.author_id,
            pr.status,
            pr.created_at,
            pr.merged_at,
            pr.version
        FROM pull_requests pr
        INNER JOIN pr_reviewers r ON pr.pull_request_id = r.pull_request_id
        WHERE r.user_id = $1
        ORDER BY pr.created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query prs by reviewer: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error or handle it appropriately
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var prs []*entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		if err := rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
			&pr.CreatedAt,
			&pr.MergedAt,
			&pr.Version,
		); err != nil {
			return nil, fmt.Errorf("scan pr: %w", err)
		}
		prs = append(prs, &pr)
	}

	return prs, rows.Err()
}

func (r *PullRequestRepository) GetOpenByReviewers(ctx context.Context, userIDs []string) ([]*entity.PullRequest, error) {
	if len(userIDs) == 0 {
		return []*entity.PullRequest{}, nil
	}

	query := `
        SELECT DISTINCT
            pr.pull_request_id,
            pr.pull_request_name,
            pr.author_id,
            pr.status,
            pr.created_at,
            pr.version,
            array_agg(r.user_id) as reviewer_ids
        FROM pull_requests pr
        INNER JOIN pr_reviewers r ON pr.pull_request_id = r.pull_request_id
        WHERE pr.status = 'OPEN' AND r.user_id = ANY($1)
        GROUP BY pr.pull_request_id
    `

	rows, err := r.db.QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, fmt.Errorf("query open prs by reviewers: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error or handle it appropriately
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var prs []*entity.PullRequest
	for rows.Next() {
		var pr entity.PullRequest
		var reviewerIDs []string

		if err := rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
			&pr.CreatedAt,
			&pr.Version,
			pq.Array(&reviewerIDs),
		); err != nil {
			return nil, fmt.Errorf("scan pr: %w", err)
		}

		pr.AssignedReviewers = reviewerIDs
		prs = append(prs, &pr)
	}

	return prs, rows.Err()
}

func (r *PullRequestRepository) AssignReviewers(ctx context.Context, prID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Используем многострочный INSERT для производительности
	valueStrings := make([]string, 0, len(userIDs))
	valueArgs := make([]any, 0, len(userIDs)*2)

	for i, userID := range userIDs {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, NOW())", i*2+1, i*2+2))
		valueArgs = append(valueArgs, prID, userID)
	}

	query := fmt.Sprintf(`
        INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at)
        VALUES %s
        ON CONFLICT (pull_request_id, user_id) DO NOTHING
    `, strings.Join(valueStrings, ","))

	_, err := r.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("assign reviewers: %w", err)
	}

	return nil
}

func (r *PullRequestRepository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	// Проверяем, назначен ли уже новый ревьюер
	isAssigned, err := r.IsReviewerAssigned(ctx, prID, newUserID)
	if err != nil {
		return fmt.Errorf("check if new reviewer is already assigned: %w", err)
	}

	if isAssigned {
		// Если новый ревьюер уже назначен, просто удаляем старого
		deleteQuery := `
            DELETE FROM pr_reviewers
            WHERE pull_request_id = $1 AND user_id = $2
        `

		result, err := r.db.ExecContext(ctx, deleteQuery, prID, oldUserID)
		if err != nil {
			return fmt.Errorf("remove old reviewer: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			return repository.ErrNotAssigned
		}

		return nil
	}

	// Атомарная операция: удаляем старого, добавляем нового
	query := `
        WITH deleted AS (
            DELETE FROM pr_reviewers
            WHERE pull_request_id = $1 AND user_id = $2
            RETURNING pull_request_id
        )
        INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at)
        SELECT $1, $3, NOW()
        WHERE EXISTS (SELECT 1 FROM deleted)
    `

	result, err := r.db.ExecContext(ctx, query, prID, oldUserID, newUserID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" { // foreign key violation
				return repository.ErrUserNotFound
			}
		}
		return fmt.Errorf("replace reviewer: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return repository.ErrNotAssigned
	}

	return nil
}

func (r *PullRequestRepository) IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM pr_reviewers
            WHERE pull_request_id = $1 AND user_id = $2
        )
    `

	var exists bool
	err := r.db.QueryRowContext(ctx, query, prID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check reviewer assigned: %w", err)
	}

	return exists, nil
}
