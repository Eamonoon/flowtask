package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"flowtask-server/internal/repository"
	"flowtask-server/internal/service"
)

func main() {
	log.Println("[CLEANUP] Starting session cleanup service...")

	// Load database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://flowtask:flowtask@localhost:5432/flowtask?sslmode=disable"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("[CLEANUP] Failed to connect to database: %v", err)
	}

	// Create repositories and services
	sessionRepo := repository.NewGenerationSessionRepository(db)
	taskRepo := repository.NewGeneratedTaskRepository(db)
	sessionService := service.NewGenerationSessionService(sessionRepo, taskRepo)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run cleanup every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("[CLEANUP] Service started, running cleanup every hour")

	// Run initial cleanup
	runCleanup(ctx, sessionService)

	for {
		select {
		case <-ticker.C:
			runCleanup(ctx, sessionService)
		case sig := <-sigChan:
			log.Printf("[CLEANUP] Received signal %v, shutting down...", sig)
			cancel()
			return
		case <-ctx.Done():
			log.Println("[CLEANUP] Context cancelled, shutting down...")
			return
		}
	}
}

func runCleanup(ctx context.Context, sessionService *service.GenerationSessionService) {
	log.Println("[CLEANUP] Running scheduled cleanup...")

	count, err := sessionService.CleanupExpiredSessions(ctx)
	if err != nil {
		log.Printf("[CLEANUP] Error during cleanup: %v", err)
		return
	}

	if count > 0 {
		log.Printf("[CLEANUP] Cleaned up %d expired sessions", count)
	} else {
		log.Println("[CLEANUP] No expired sessions found")
	}

	// Log current stats
	stats, err := sessionService.GetSessionStats(ctx)
	if err != nil {
		log.Printf("[CLEANUP] Error getting stats: %v", err)
		return
	}

	log.Printf("[CLEANUP] Current stats - generating: %d, completed: %d, expired: %d",
		stats["generating"], stats["completed"], stats["expired"])
}
