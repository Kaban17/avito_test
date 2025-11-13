package entity

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
	Version           int
}

func (pr *PullRequest) HasReviewer(userID string) bool {
	for _, id := range pr.AssignedReviewers {
		if id == userID {
			return true
		}
	}
	return false
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == StatusMerged
}
