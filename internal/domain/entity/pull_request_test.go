package entity

import (
	"testing"
)

func TestPullRequest_HasReviewer(t *testing.T) {
	pr := &PullRequest{
		AssignedReviewers: []string{"user1", "user2", "user3"},
	}

	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{"User in reviewers", "user2", true},
		{"User not in reviewers", "user4", false},
		{"Empty user ID", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pr.HasReviewer(tt.userID)
			if result != tt.expected {
				t.Errorf("HasReviewer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPullRequest_IsMerged(t *testing.T) {
	tests := []struct {
		name     string
		status   PRStatus
		expected bool
	}{
		{"Merged PR", StatusMerged, true},
		{"Open PR", StatusOpen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PullRequest{Status: tt.status}
			result := pr.IsMerged()
			if result != tt.expected {
				t.Errorf("IsMerged() = %v, want %v", result, tt.expected)
			}
		})
	}
}
