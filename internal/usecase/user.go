package usecase

import (
	"context"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"
)

type UserUseCase struct {
	txManager repository.TxManager
}

func NewUserUseCase(txManager repository.TxManager) *UserUseCase {
	return &UserUseCase{txManager: txManager}
}

func (uc *UserUseCase) SetActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {
	var result *entity.User

	err := uc.txManager.WithTx(ctx, func(tx repository.Tx) error {
		if err := tx.Users().SetActive(ctx, userID, isActive); err != nil {
			return err
		}

		user, err := tx.Users().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		result = user
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (uc *UserUseCase) GetReviews(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	var result []*entity.PullRequest

	err := uc.txManager.WithTx(ctx, func(tx repository.Tx) error {
		_, err := tx.Users().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		prs, err := tx.PullRequests().GetByReviewer(ctx, userID)
		if err != nil {
			return err
		}

		result = prs
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
