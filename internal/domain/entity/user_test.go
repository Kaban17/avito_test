package entity

import (
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	now := time.Now()

	user := &User{
		UserID:    "user123",
		Username:  "john_doe",
		TeamName:  "backend",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Проверяем, что поля установлены корректно
	if user.UserID != "user123" {
		t.Errorf("Expected UserID to be 'user123', got %s", user.UserID)
	}

	if user.Username != "john_doe" {
		t.Errorf("Expected Username to be 'john_doe', got %s", user.Username)
	}

	if user.TeamName != "backend" {
		t.Errorf("Expected TeamName to be 'backend', got %s", user.TeamName)
	}

	if !user.IsActive {
		t.Error("Expected IsActive to be true")
	}

	if user.CreatedAt != now {
		t.Error("CreatedAt not set correctly")
	}

	if user.UpdatedAt != now {
		t.Error("UpdatedAt not set correctly")
	}
}
