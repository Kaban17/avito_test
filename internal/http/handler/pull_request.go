// internal/http/handler/pull_request.go
package handler

import (
	"encoding/json"
	"net/http"

	"reviewer-service/internal/http/response"
	"reviewer-service/internal/repository"
	"reviewer-service/internal/usecase"
)

type PullRequestHandler struct {
	prUC *usecase.PullRequestUseCase
}

func NewPullRequestHandler(prUC *usecase.PullRequestUseCase) *PullRequestHandler {
	return &PullRequestHandler{prUC: prUC}
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

func (h *PullRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Валидация
	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "all fields are required")
		return
	}

	pr, err := h.prUC.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		if err == repository.ErrPRExists {
			response.Error(w, http.StatusConflict, "PR_EXISTS", "PR id already exists")
			return
		}
		if err == repository.ErrNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "author not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"pr": pr,
	})
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

func (h *PullRequestHandler) Merge(w http.ResponseWriter, r *http.Request) {
	var req MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.PullRequestID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "pull_request_id is required")
		return
	}

	pr, err := h.prUC.Merge(r.Context(), req.PullRequestID)
	if err != nil {
		if err == repository.ErrNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "pull request not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"pr": pr,
	})
}

type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

func (h *PullRequestHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "all fields are required")
		return
	}

	pr, newReviewerID, err := h.prUC.Reassign(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		if err == repository.ErrPRMerged {
			response.Error(w, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
			return
		}
		if err == repository.ErrNotAssigned {
			response.Error(w, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
			return
		}
		if err == repository.ErrNoCandidate {
			response.Error(w, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
			return
		}
		if err == repository.ErrNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "PR or user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
