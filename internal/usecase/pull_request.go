package usecase

import (
	"context"
	"fmt"
	"time"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/domain/service"
	"reviewer-service/internal/repository"
	"reviewer-service/internal/repository/postgres"
)

type PullRequestUseCase struct {
	txManager *postgres.TxManager
	selector  *service.ReviewerSelector
}

func NewPullRequestUseCase(
	txManager *postgres.TxManager,
	selector *service.ReviewerSelector,
) *PullRequestUseCase {
	return &PullRequestUseCase{
		txManager: txManager,
		selector:  selector,
	}
}

// CreatePR создаёт PR и назначает ревьюверов атомарно
func (uc *PullRequestUseCase) CreatePR(
	ctx context.Context,
	prID, prName, authorID string,
) (*entity.PullRequest, error) {
	var result *entity.PullRequest

	err := uc.txManager.WithTx(ctx, func(tx repository.Tx) error {
		// 1. Получаем автора
		author, err := tx.Users().GetByID(ctx, authorID)
		if err != nil {
			return fmt.Errorf("get author: %w", err)
		}

		// 2. Создаём PR
		pr := &entity.PullRequest{
			ID:       prID,
			Name:     prName,
			AuthorID: authorID,
			Status:   entity.StatusOpen,
		}

		if err := tx.PullRequests().Create(ctx, pr); err != nil {
			return fmt.Errorf("create pr: %w", err)
		}

		// 3. Выбираем ревьюверов (передаём tx!)
		reviewers, err := uc.selector.Select(ctx, tx, author.TeamName, authorID)
		if err != nil {
			return fmt.Errorf("select reviewers: %w", err)
		}

		// 4. Назначаем их
		if len(reviewers) > 0 {
			reviewerIDs := make([]string, len(reviewers))
			for i, r := range reviewers {
				reviewerIDs[i] = r.UserID
			}

			if err := tx.PullRequests().AssignReviewers(ctx, prID, reviewerIDs); err != nil {
				return fmt.Errorf("assign reviewers: %w", err)
			}

			// Обновляем статистику
			for _, reviewerID := range reviewerIDs {
				if err := tx.Stats().IncrementAssignment(ctx, reviewerID); err != nil {
					return fmt.Errorf("increment assignment: %w", err)
				}
			}

			pr.AssignedReviewers = reviewerIDs
		}

		result = pr
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Merge помечает PR как merged (идемпотентная операция)
func (uc *PullRequestUseCase) Merge(
	ctx context.Context,
	prID string,
) (*entity.PullRequest, error) {
	var result *entity.PullRequest

	err := uc.txManager.WithTx(ctx, func(tx repository.Tx) error {
		// Получаем PR с блокировкой
		pr, err := tx.PullRequests().GetByIDForUpdate(ctx, prID)
		if err != nil {
			return err
		}

		// Уже merged? Возвращаем как есть (идемпотентность!)
		if pr.Status == entity.StatusMerged {
			result = pr
			return nil
		}

		// Merging
		pr.Status = entity.StatusMerged
		now := time.Now()
		pr.MergedAt = &now

		if err := tx.PullRequests().Update(ctx, pr); err != nil {
			return err
		}

		result = pr
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Reassign переназначает ревьювера
func (uc *PullRequestUseCase) Reassign(
	ctx context.Context,
	prID, oldUserID string,
) (*entity.PullRequest, string, error) {
	var result *entity.PullRequest
	var newReviewerID string

	err := uc.txManager.WithTx(ctx, func(tx repository.Tx) error {
		// 1. Проверяем PR
		pr, err := tx.PullRequests().GetByIDForUpdate(ctx, prID)
		if err != nil {
			return err
		}

		if pr.Status == entity.StatusMerged {
			return repository.ErrPRMerged
		}

		// 2. Проверяем что oldUser назначен
		isAssigned, err := tx.PullRequests().IsReviewerAssigned(ctx, prID, oldUserID)
		if err != nil {
			return err
		}
		if !isAssigned {
			return repository.ErrNotAssigned
		}

		// 3. Получаем старого ревьювера для определения команды
		oldUser, err := tx.Users().GetByID(ctx, oldUserID)
		if err != nil {
			return err
		}

		// 4. Выбираем замену из ЕГО команды (передаём tx!)
		newReviewer, err := uc.selector.SelectReplacement(
			ctx,
			tx,
			oldUser.TeamName,
			[]string{oldUserID, pr.AuthorID},
		)
		if err != nil {
			return err
		}

		if newReviewer == nil {
			return repository.ErrNoCandidate
		}

		// 5. Атомарная замена
		if err := tx.PullRequests().ReplaceReviewer(ctx, prID, oldUserID, newReviewer.UserID); err != nil {
			return err
		}

		// 6. Обновляем статистику
		if err := tx.Stats().IncrementAssignment(ctx, newReviewer.UserID); err != nil {
			return err
		}

		// Обновляем PR объект
		for i, id := range pr.AssignedReviewers {
			if id == oldUserID {
				pr.AssignedReviewers[i] = newReviewer.UserID
				break
			}
		}

		result = pr
		newReviewerID = newReviewer.UserID
		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return result, newReviewerID, nil
}
