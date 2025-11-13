package entity

import "time"

type PullRequest struct {
	ID                int64
	Name              string
	AuthorID          string
	AssignedReviewers []string
	State             string
	Status            string
	CreatedAt         time.Time
	MergedAt          time.Time
	Version           int64
}
