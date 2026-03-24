package handler

import (
	"net/http"

	"github.com/2hot4you/aiaos/backend/internal/middleware"
	"github.com/2hot4you/aiaos/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(
	authSvc *service.AuthService,
	adminSvc *service.AdminService,
	projectSvc *service.ProjectService,
	seasonSvc *service.SeasonService,
	episodeSvc *service.EpisodeService,
) *Router {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Handlers
	authH := NewAuthHandler(authSvc)
	adminH := NewAdminHandler(adminSvc)
	projectH := NewProjectHandler(projectSvc)
	seasonH := NewSeasonHandler(seasonSvc)
	episodeH := NewEpisodeHandler(episodeSvc)

	api := r.Group("/api/v1")

	// Public routes
	api.POST("/auth/login", authH.Login)

	// Authenticated routes
	auth := api.Group("")
	auth.Use(middleware.Auth(authSvc))
	{
		// Auth
		auth.GET("/auth/me", authH.GetMe)
		auth.PUT("/users/me/password", authH.ChangePassword)

		// Projects
		auth.GET("/projects", projectH.List)
		auth.POST("/projects", projectH.Create)
		auth.GET("/projects/:id", projectH.Get)
		auth.PUT("/projects/:id", projectH.Update)
		auth.DELETE("/projects/:id", projectH.Delete)

		// Seasons
		auth.GET("/projects/:id/seasons", seasonH.ListByProject)
		auth.POST("/projects/:id/seasons", seasonH.Create)
		auth.PUT("/seasons/:id", seasonH.Update)
		auth.DELETE("/seasons/:id", seasonH.Delete)

		// Episodes
		auth.GET("/seasons/:sid/episodes", episodeH.ListBySeason)
		auth.POST("/seasons/:sid/episodes", episodeH.Create)
		auth.GET("/episodes/:id", episodeH.Get)
		auth.PUT("/episodes/:id", episodeH.Update)
		auth.DELETE("/episodes/:id", episodeH.Delete)

		// Admin routes
		admin := auth.Group("/admin")
		admin.Use(middleware.RequireAdmin())
		{
			// User management
			admin.GET("/users", adminH.ListUsers)
			admin.POST("/users", adminH.CreateUser)
			admin.PUT("/users/:id", adminH.UpdateUser)
			admin.DELETE("/users/:id", adminH.DeleteUser)
			admin.POST("/users/:id/reset-password", adminH.ResetPassword)
			admin.PUT("/users/:id/status", adminH.UpdateUserStatus)

			// Model management
			admin.GET("/models", adminH.ListModels)
			admin.POST("/models", adminH.CreateModel)
			admin.PUT("/models/:id", adminH.UpdateModel)
			admin.DELETE("/models/:id", adminH.DeleteModel)
		}
	}

	return &Router{engine: r}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
