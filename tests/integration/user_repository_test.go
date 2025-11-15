package integration

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/repository/postgres"
)

// getTestDBURL получает URL для подключения к тестовой БД из переменных окружения
func getTestDBURL() string {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		// URL по умолчанию для локального тестирования
		return "postgres://postgres:postgres@localhost:5432/reviewer_service_test?sslmode=disable"
	}
	return url
}

// setupTestDB создает подключение к тестовой БД
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", getTestDBURL())
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Очищаем таблицы перед каждым тестом
	_, err = db.Exec(`
		DELETE FROM pull_request_reviewers;
		DELETE FROM pull_requests;
		DELETE FROM users;
		DELETE FROM teams;
	`)
	if err != nil {
		t.Fatalf("Failed to clean test database: %v", err)
	}

	return db
}

// TestUserRepository_Integration тестирует основные операции с пользователями
func TestUserRepository_Integration(t *testing.T) {
	// Пропускаем тест, если не задана переменная окружения для тестовой БД
	if os.Getenv("INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test - set INTEGRATION_TESTS=1 to run")
	}

	db := setupTestDB(t)
	defer db.Close()

	// Создаем репозиторий
	repo := postgres.NewUserRepository(db)

	ctx := context.Background()

	// Создаем тестовую команду
	teamName := "integration_test_team_" + uuid.New().String()
	_, err := db.Exec(
		"INSERT INTO teams (name, created_at) VALUES ($1, $2)",
		teamName,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to create test team: %v", err)
	}

	// Тест 1: Создание пользователя
	t.Run("CreateUser", func(t *testing.T) {
		user := &entity.User{
			UserID:    "test_user_" + uuid.New().String(),
			Username:  "testuser",
			TeamName:  teamName,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Проверяем, что пользователь был создан
		createdUser, err := repo.GetByID(ctx, user.UserID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if createdUser.UserID != user.UserID {
			t.Errorf("GetByID() UserID = %v, want %v", createdUser.UserID, user.UserID)
		}

		if createdUser.Username != user.Username {
			t.Errorf("GetByID() Username = %v, want %v", createdUser.Username, user.Username)
		}

		if createdUser.TeamName != user.TeamName {
			t.Errorf("GetByID() TeamName = %v, want %v", createdUser.TeamName, user.TeamName)
		}

		if createdUser.IsActive != user.IsActive {
			t.Errorf("GetByID() IsActive = %v, want %v", createdUser.IsActive, user.IsActive)
		}
	})

	// Тест 2: Получение пользователя по ID
	t.Run("GetUserByID", func(t *testing.T) {
		// Создаем пользователя для теста
		userID := "get_by_id_user_" + uuid.New().String()
		user := &entity.User{
			UserID:    userID,
			Username:  "getbyiduser",
			TeamName:  teamName,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Получаем пользователя
		fetchedUser, err := repo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if fetchedUser.UserID != userID {
			t.Errorf("GetByID() UserID = %v, want %v", fetchedUser.UserID, userID)
		}
	})

	// Тест 3: Установка статуса активности
	t.Run("SetUserActive", func(t *testing.T) {
		// Создаем пользователя для теста
		userID := "set_active_user_" + uuid.New().String()
		user := &entity.User{
			UserID:    userID,
			Username:  "setactiveuser",
			TeamName:  teamName,
			IsActive:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Устанавливаем статус активности
		err = repo.SetActive(ctx, userID, true)
		if err != nil {
			t.Fatalf("SetActive() error = %v", err)
		}

		// Проверяем, что статус изменился
		updatedUser, err := repo.GetByID(ctx, userID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if !updatedUser.IsActive {
			t.Error("SetActive() failed to update IsActive status")
		}
	})

	// Тест 4: Получение пользователей по команде
	t.Run("GetUsersByTeam", func(t *testing.T) {
		// Создаем дополнительную команду для теста
		otherTeamName := "other_team_" + uuid.New().String()
		_, err := db.Exec(
			"INSERT INTO teams (name, created_at) VALUES ($1, $2)",
			otherTeamName,
			time.Now(),
		)
		if err != nil {
			t.Fatalf("Failed to create other team: %v", err)
		}

		// Создаем пользователей для теста
		users := []*entity.User{
			{
				UserID:    "team_user_1_" + uuid.New().String(),
				Username:  "teamuser1",
				TeamName:  teamName,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				UserID:    "team_user_2_" + uuid.New().String(),
				Username:  "teamuser2",
				TeamName:  teamName,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				UserID:    "other_team_user_" + uuid.New().String(),
				Username:  "otherteamuser",
				TeamName:  otherTeamName,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		// Сохраняем пользователей
		for _, user := range users {
			err := repo.Create(ctx, user)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
		}

		// Получаем пользователей по команде
		teamUsers, err := repo.GetByTeam(ctx, teamName)
		if err != nil {
			t.Fatalf("GetByTeam() error = %v", err)
		}

		// Проверяем, что получили правильное количество пользователей
		if len(teamUsers) != 2 {
			t.Errorf("GetByTeam() returned %v users, want 2", len(teamUsers))
		}

		// Проверяем, что все пользователи принадлежат правильной команде
		for _, user := range teamUsers {
			if user.TeamName != teamName {
				t.Errorf("GetByTeam() returned user with TeamName = %v, want %v", user.TeamName, teamName)
			}
		}
	})
}
