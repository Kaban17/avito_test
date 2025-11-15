package entity

import (
	"testing"
	"time"
)

func TestTeam(t *testing.T) {
	now := time.Now()

	// Создаем пользователей для команды
	users := []*User{
		{
			UserID:   "user1",
			Username: "alice",
			TeamName: "backend",
			IsActive: true,
		},
		{
			UserID:   "user2",
			Username: "bob",
			TeamName: "backend",
			IsActive: false,
		},
	}

	team := &Team{
		Name:      "backend",
		Members:   users,
		CreatedAt: now,
	}

	// Проверяем, что поля установлены корректно
	if team.Name != "backend" {
		t.Errorf("Expected Name to be 'backend', got %s", team.Name)
	}

	if len(team.Members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(team.Members))
	}

	if team.CreatedAt != now {
		t.Error("CreatedAt not set correctly")
	}

	// Проверяем членов команды
	if team.Members[0].UserID != "user1" {
		t.Errorf("Expected first member ID to be 'user1', got %s", team.Members[0].UserID)
	}

	if team.Members[1].Username != "bob" {
		t.Errorf("Expected second member username to be 'bob', got %s", team.Members[1].Username)
	}
}
