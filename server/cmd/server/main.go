package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"flowtask-server/internal/ai"
	"flowtask-server/internal/config"
	"flowtask-server/internal/handler"
	"flowtask-server/internal/middleware"
	"flowtask-server/internal/repository"
	"flowtask-server/internal/service"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := config.NewDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := config.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	rdb, err := config.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	aiClient := ai.NewClient(cfg.AI)

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// Repositories
	userRepo := repository.NewUserRepository(db)
	goalRepo := repository.NewLearningGoalRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	labelRepo := repository.NewLabelRepository(db)
	sessionRepo := repository.NewStudySessionRepository(db)
	convRepo := repository.NewAIConversationRepository(db)
	genSessionRepo := repository.NewGenerationSessionRepository(db, rdb)
	genTaskRepo := repository.NewGeneratedTaskRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, rdb, cfg.JWT)
	genSessionService := service.NewGenerationSessionService(genSessionRepo, genTaskRepo)
	goalService := service.NewLearningGoalService(goalRepo, taskRepo, aiClient, genSessionService)
	labelService := service.NewLabelService(labelRepo)
	dashboardService := service.NewDashboardService(taskRepo, sessionRepo, rdb)
	taskService := service.NewTaskService(taskRepo, dashboardService)
	chatService := service.NewAIChatService(convRepo, taskRepo, goalRepo, sessionRepo, aiClient)
	sessionService := service.NewStudySessionService(sessionRepo)
	userService := service.NewUserService(userRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	goalHandler := handler.NewLearningGoalHandler(goalService)
	taskHandler := handler.NewTaskHandler(taskService)
	labelHandler := handler.NewLabelHandler(labelService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	chatHandler := handler.NewAIChatHandler(chatService)
	sessionHandler := handler.NewStudySessionHandler(sessionService)
	userHandler := handler.NewUserHandler(userService)

	api := r.Group("/api")
	{
		// Health
		healthHandler := handler.NewHealthHandler(db, rdb, genSessionService)
		api.GET("/health", healthHandler.Check)

		// Auth (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.JWT.AccessSecret))
		{
			// User
			protected.GET("/user/profile", userHandler.GetProfile)
			protected.PUT("/user/profile", userHandler.UpdateProfile)

			// Learning Goals
			protected.POST("/learning-goals", goalHandler.CreateWithSession)
			protected.GET("/learning-goals", goalHandler.List)
			protected.GET("/learning-goals/:id", goalHandler.Get)
			protected.PUT("/learning-goals/:id", goalHandler.Update)
			protected.POST("/learning-goals/:id/tasks", goalHandler.AddTask)
			protected.DELETE("/learning-goals/:id/tasks/:taskId", goalHandler.DeleteTask)
			protected.GET("/learning-goals/:id/generate-stream", goalHandler.GenerateStream)
			protected.POST("/learning-goals/:id/tasks/confirm", goalHandler.ConfirmTasks)
			protected.POST("/learning-goals/:id/regenerate", goalHandler.Regenerate)

			// Tasks
			protected.POST("/tasks", taskHandler.Create)
			protected.GET("/tasks", taskHandler.List)
			protected.GET("/tasks/:id", taskHandler.Get)
			protected.PUT("/tasks/:id", taskHandler.Update)
			protected.DELETE("/tasks/:id", taskHandler.Delete)
			protected.POST("/tasks/:id/dependencies", taskHandler.AddDependency)
				protected.GET("/tasks/:id/subtasks", taskHandler.ListSubtasks)
				protected.POST("/tasks/:id/subtasks", taskHandler.CreateSubtask)
				protected.PATCH("/tasks/:id/subtasks/:subtaskId", taskHandler.UpdateSubtask)
				protected.DELETE("/tasks/:id/subtasks/:subtaskId", taskHandler.DeleteSubtask)

			// Labels
			protected.POST("/labels", labelHandler.Create)
			protected.GET("/labels", labelHandler.List)
			protected.DELETE("/labels/:id", labelHandler.Delete)

			// Dashboard
			protected.GET("/dashboard/stats", dashboardHandler.GetStats)
			protected.GET("/dashboard/charts/study-time", dashboardHandler.GetStudyTimeChart)
			protected.GET("/dashboard/charts/category-stats", dashboardHandler.GetCategoryStats)
			protected.GET("/dashboard/charts/completion-rate", dashboardHandler.GetCompletionRateChart)

			// AI Chat
			protected.POST("/ai/chat", chatHandler.Chat)
			protected.GET("/ai/conversations", chatHandler.ListConversations)
			protected.GET("/ai/conversations/:id/messages", chatHandler.GetMessages)
				protected.DELETE("/ai/conversations/:id", chatHandler.DeleteConversation)

			// Study Sessions
			protected.POST("/study-sessions", sessionHandler.Create)
			protected.GET("/study-sessions", sessionHandler.List)
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
