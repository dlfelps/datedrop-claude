package api

import (
	"github.com/gin-gonic/gin"

	"datedrop/internal/api/handlers"
	"datedrop/internal/api/middleware"
)

type Router struct {
	userHandler       *handlers.UserHandler
	quizHandler       *handlers.QuizHandler
	matchingHandler   *handlers.MatchingHandler
	dropHandler       *handlers.DropHandler
	socialHandler     *handlers.SocialHandler
	moderationHandler *handlers.ModerationHandler
	seedHandler       *handlers.SeedHandler
}

func NewRouter(
	userHandler *handlers.UserHandler,
	quizHandler *handlers.QuizHandler,
	matchingHandler *handlers.MatchingHandler,
	dropHandler *handlers.DropHandler,
	socialHandler *handlers.SocialHandler,
	moderationHandler *handlers.ModerationHandler,
	seedHandler *handlers.SeedHandler,
) *Router {
	return &Router{
		userHandler:       userHandler,
		quizHandler:       quizHandler,
		matchingHandler:   matchingHandler,
		dropHandler:       dropHandler,
		socialHandler:     socialHandler,
		moderationHandler: moderationHandler,
		seedHandler:       seedHandler,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public endpoints (no auth)
	api := engine.Group("/api/v1")
	{
		api.POST("/users", r.userHandler.CreateUser)
		api.POST("/auth/login", r.userHandler.Login)
	}

	// Protected endpoints (require auth)
	protected := engine.Group("/api/v1")
	protected.Use(middleware.MockAuth())
	{
		// User
		protected.GET("/users/:id", r.userHandler.GetUser)
		protected.PATCH("/users/:id", r.userHandler.UpdateUser)

		// Quiz
		protected.GET("/quiz/questions", r.quizHandler.GetQuestions)
		protected.POST("/quiz/responses", r.quizHandler.SubmitResponse)
		protected.GET("/quiz/status/:user_id", r.quizHandler.GetStatus)

		// Matching
		protected.POST("/matching/run", r.matchingHandler.RunMatching)

		// Drops
		protected.GET("/drops/current", r.dropHandler.GetCurrentDrop)
		protected.POST("/drops/:id/accept", r.dropHandler.AcceptDrop)
		protected.POST("/drops/:id/decline", r.dropHandler.DeclineDrop)
		protected.GET("/drops/history", r.dropHandler.GetDropHistory)

		// Social
		protected.GET("/social/users", r.socialHandler.BrowseUsers)
		protected.POST("/social/shoot", r.socialHandler.ShootYourShot)
		protected.GET("/social/shots/mutual", r.socialHandler.GetMutualShots)
		protected.POST("/social/cupid", r.socialHandler.NominateCupid)
		protected.POST("/social/cupid/:id/accept", r.socialHandler.AcceptCupid)
		protected.POST("/social/cupid/:id/decline", r.socialHandler.DeclineCupid)

		// Moderation
		protected.POST("/moderation/block/:user_id", r.moderationHandler.BlockUser)
		protected.DELETE("/moderation/block/:user_id", r.moderationHandler.UnblockUser)
		protected.POST("/moderation/report", r.moderationHandler.ReportUser)
	}

	// Debug endpoints
	debug := engine.Group("/debug")
	{
		debug.POST("/seed", r.seedHandler.Seed)
	}
}
