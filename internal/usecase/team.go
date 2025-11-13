package usecase

import (
	"context"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"
	"reviewer-service/internal/repository/postgres"
)

type TeamUseCase struct {
	txManager *postgres.TxManager
}

func NewTeamUseCase(txManager *postgres.TxManager) *TeamUseCase {
	return &TeamUseCase{txManager: txManager}
}

func (uc *TeamUseCase) CreateTeam(ctx context.Context, team *entity.Team) (*entity.Team, error) {
	err := uc.txManager.WithTx(ctx, func(tx repository.Tx) error {
		// Проверяем существование
		exists, err := tx.Teams().Exists(ctx, team.Name)
		if err != nil {
			return err
		}
		if exists {
			return repository.ErrTeamExists
		}

		return tx.Teams().Create(ctx, team)
	})

	if err != nil {
		return nil, err
	}

	return team, nil
}

func (uc *TeamUseCase) GetTeam(ctx context.Context, name string) (*entity.Team, error) {
	var result *entity.Team

	err := uc.txManager.WithTx(ctx, func(tx repository.Tx) error {
		team, err := tx.Teams().GetByName(ctx, name)
		if err != nil {
			return err
		}
		result = team
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
