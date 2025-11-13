// internal/domain/service/reviewer_selector.go
package service

import (
	"context"
	"math/rand/v2"
	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository"
	"sort"
)

type ReviewerSelector struct {
	userRepo  repository.UserRepository
	statsRepo repository.StatsRepository
}

// Select выбирает до 2 ревьюеров с учётом fair distribution
func (s *ReviewerSelector) Select(
	ctx context.Context,
	teamName string,
	authorID string,
) ([]*entity.User, error) {
	// 1. Получаем активных участников команды (кроме автора)
	candidates, err := s.userRepo.GetActiveByTeam(ctx, teamName, authorID)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return []*entity.User{}, nil
	}

	// 2. Получаем текущую нагрузку каждого
	workload, err := s.statsRepo.GetWorkload(ctx, candidateIDs(candidates))
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
func candidateIDs(users []*entity.User) []string {
	var ids []string
	for _, user := range users {
		ids = append(ids, user.UserID)
	}
	return ids
}
