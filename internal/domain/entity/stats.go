package entity

import "time"

type UserStats struct {
	ID              int64
	UserID          string
	Username        string
	TeamName        string
	OpenReviews     int64
	TotalReviews    int64
	AssignmentCount int64
	LastAssignedAt  time.Time
	CreatedAt       time.Time
}

type TeamStats struct {
	TeamID        string `json:"team_id"`
	TeamName      string `json:"team_name"`
	OpenPRs       int    `json:"open_prs"`
	OpenReviews   int    `json:"open_reviews"`
	TotalMembers  int    `json:"total_members"`
	ActiveMembers int    `json:"active_members"`
	TotalPRs      int    `json:"total_prs"`
}
