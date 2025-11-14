// internal/http/handler/user.go
package handler

import (
	"encoding/json"
	"net/http"

	"reviewer-service/internal/http/response"
	"reviewer-service/internal/repository"
	"reviewer-service/internal/usecase"
)

type UserHandler struct {
	userUC *usecase.UserUseCase
}

func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

type SetActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.UserID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	user, err := h.userUC.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		if err == repository.ErrNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id query parameter is required")
		return
	}

	prs, err := h.userUC.GetReviews(r.Context(), userID)
	if err != nil {
		if err == repository.ErrNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
