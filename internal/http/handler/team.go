package handler

import (
	"encoding/json"
	"net/http"

	"reviewer-service/internal/domain/entity"
	"reviewer-service/internal/http/response"
	"reviewer-service/internal/repository"
	"reviewer-service/internal/usecase"
)

type TeamHandler struct {
	teamUC *usecase.TeamUseCase
}

func NewTeamHandler(teamUC *usecase.TeamUseCase) *TeamHandler {
	return &TeamHandler{teamUC: teamUC}
}

type CreateTeamRequest struct {
	TeamName string                `json:"team_name"`
	Members  []CreateMemberRequest `json:"members"`
}

type CreateMemberRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Валидация
	if req.TeamName == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	// Конвертируем в entity
	team := &entity.Team{
		Name:    req.TeamName,
		Members: make([]*entity.User, len(req.Members)),
	}

	for i, m := range req.Members {
		team.Members[i] = &entity.User{
			UserID:   m.UserID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		}
	}

	// Создаём команду
	result, err := h.teamUC.CreateTeam(r.Context(), team)
	if err != nil {
		if err == repository.ErrTeamExists {
			response.Error(w, http.StatusBadRequest, "TEAM_EXISTS", "team_name already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"team": result,
	})
}

func (h *TeamHandler) Get(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name query parameter is required")
		return
	}

	team, err := h.teamUC.GetTeam(r.Context(), teamName)
	if err != nil {
		if err == repository.ErrNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "team not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, team)
}
