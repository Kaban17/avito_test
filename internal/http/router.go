// internal/http/router.go
package http

import (
	"net/http"

	"reviewer-service/internal/http/handler"
	"reviewer-service/internal/http/middleware"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	teamHandler *handler.TeamHandler
	userHandler *handler.UserHandler
	prHandler   *handler.PullRequestHandler
}

func NewRouter(
	teamHandler *handler.TeamHandler,
	userHandler *handler.UserHandler,
	prHandler *handler.PullRequestHandler,
) *Router {
	return &Router{
		teamHandler: teamHandler,
		userHandler: userHandler,
		prHandler:   prHandler,
	}
}

func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.RequestID)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Teams
	r.Post("/team/add", rt.teamHandler.Create)
	r.Get("/team/get", rt.teamHandler.Get)

	// Users
	r.Post("/users/setIsActive", rt.userHandler.SetIsActive)
	r.Get("/users/getReview", rt.userHandler.GetReview)

	// Pull Requests
	r.Post("/pullRequest/create", rt.prHandler.Create)
	r.Post("/pullRequest/merge", rt.prHandler.Merge)
	r.Post("/pullRequest/reassign", rt.prHandler.Reassign)

	return r
}
