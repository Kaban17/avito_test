package service

import (
	"context"
	"math/rand"
	"sort"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"
)

type ReviewerSelector struct {
	// Без зависимостей - работаем через tx
}

func NewReviewerSelector() *ReviewerSelector {
	return &ReviewerSelector{}
}

// Select выбирает до 2 ревьюеров с учётом fair distribution
func (s *ReviewerSelector) Select(
	ctx context.Context,
	tx repository.Tx,
	teamName string,
	authorID string,
) ([]*entity.User, error) {
	// 1. Получаем активных участников команды (кроме автора)
	candidates, err := tx.Users().GetActiveByTeam(ctx, teamName, authorID)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return []*entity.User{}, nil
	}

	// 2. Получаем текущую нагрузку каждого
	candidateIDs := make([]string, len(candidates))
	for i, c := range candidates {
		candidateIDs[i] = c.UserID
	}

	workload, err := tx.Stats().GetWorkload(ctx, candidateIDs)
	if err != nil {
		return nil, err
	}

	// 3. Сортируем по нагрузке (меньше нагрузки = выше приоритет)
	sort.Slice(candidates, func(i, j int) bool {
		loadI := workload[candidates[i].UserID]
		loadJ := workload[candidates[j].UserID]

		if loadI == loadJ {
			// При равной нагрузке - случайный выбор
			return rand.Float32() > 0.5
		}
		return loadI < loadJ
	})

	// 4. Берём топ-2
	count := min(2, len(candidates))
	return candidates[:count], nil
}

// SelectReplacement выбирает одного ревьювера на замену
func (s *ReviewerSelector) SelectReplacement(
	ctx context.Context,
	tx repository.Tx,
	teamName string,
	excludeUserIDs []string,
) (*entity.User, error) {
	// Получаем всех активных из команды
	allUsers, err := tx.Users().GetByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	// Фильтруем: только активные + не в exclude списке
	excludeMap := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeMap[id] = true
	}

	var candidates []*entity.User
	var candidateIDs []string

	for _, user := range allUsers {
		if user.IsActive && !excludeMap[user.UserID] {
			candidates = append(candidates, user)
			candidateIDs = append(candidateIDs, user.UserID)
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// Получаем нагрузку
	workload, err := tx.Stats().GetWorkload(ctx, candidateIDs)
	if err != nil {
		return nil, err
	}

	// Сортируем по нагрузке
	sort.Slice(candidates, func(i, j int) bool {
		loadI := workload[candidates[i].UserID]
		loadJ := workload[candidates[j].UserID]

		if loadI == loadJ {
			return rand.Float32() > 0.5
		}
		return loadI < loadJ
	})

	// Возвращаем первого (с минимальной нагрузкой)
	return candidates[0], nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
