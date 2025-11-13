package repository

import "errors"

var (
	// Generic errors
	ErrNotFound       = errors.New("resource not found")
	ErrOptimisticLock = errors.New("optimistic lock failure")

	// Team errors
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")

	// User errors
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	// PR errors
	ErrPRExists   = errors.New("pull request already exists")
	ErrPRNotFound = errors.New("pull request not found")
	ErrPRMerged   = errors.New("pull request is merged")

	// Reviewer errors
	ErrNotAssigned = errors.New("reviewer not assigned to this PR")
	ErrNoCandidate = errors.New("no candidate available for assignment")
)
