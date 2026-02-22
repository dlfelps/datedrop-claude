package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"datedrop/internal/api"
	"datedrop/internal/api/handlers"
	"datedrop/internal/config"
	"datedrop/internal/repository/memory"
	"datedrop/internal/services"
)

func main() {
	cfg := config.NewDefaultConfig()

	// Repositories
	userRepo := memory.NewUserRepository()
	questionRepo := memory.NewQuestionRepository()
	responseRepo := memory.NewResponseRepository()
	dropRepo := memory.NewDropRepository()
	shotRepo := memory.NewShotRepository()
	cupidRepo := memory.NewCupidRepository()
	moderationRepo := memory.NewModerationRepository()

	// Services
	notificationService := services.NewNotificationService()
	userService := services.NewUserService(userRepo)
	quizService := services.NewQuizService(questionRepo, responseRepo, userRepo, cfg.Quiz.TotalQuestions)
	matchingService := services.NewMatchingService(cfg, userRepo, questionRepo, responseRepo, dropRepo, moderationRepo, notificationService)
	dropService := services.NewDropService(dropRepo, notificationService)
	socialService := services.NewSocialService(cfg, shotRepo, cupidRepo, userRepo, dropRepo, moderationRepo, matchingService, notificationService)
	moderationService := services.NewModerationService(moderationRepo, notificationService)

	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	quizHandler := handlers.NewQuizHandler(quizService)
	matchingHandler := handlers.NewMatchingHandler(matchingService)
	dropHandler := handlers.NewDropHandler(dropService)
	socialHandler := handlers.NewSocialHandler(socialService)
	moderationHandler := handlers.NewModerationHandler(moderationService)
	seedHandler := handlers.NewSeedHandler(userRepo, questionRepo, responseRepo)

	// Router
	router := api.NewRouter(
		userHandler,
		quizHandler,
		matchingHandler,
		dropHandler,
		socialHandler,
		moderationHandler,
		seedHandler,
	)

	engine := gin.Default()
	router.Setup(engine)

	log.Printf("Starting DateDrop server on %s", cfg.Server.Port)
	if err := engine.Run(cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
