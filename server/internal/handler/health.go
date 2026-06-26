package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"flowtask-server/internal/response"
	"flowtask-server/internal/service"
)

type HealthHandler struct {
	db              *gorm.DB
	redis           *redis.Client
	sessionService  *service.GenerationSessionService
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client, sessionService *service.GenerationSessionService) *HealthHandler {
	return &HealthHandler{
		db:              db,
		redis:           redis,
		sessionService:  sessionService,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	data := gin.H{
		"status":     "ok",
		"database":   "ok",
		"redis":      "ok",
		"ai_service": "ok",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		data["database"] = "error"
		data["status"] = "degraded"
	}

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		data["redis"] = "error"
		data["status"] = "degraded"
	}

	// Get session metrics
	if h.sessionService != nil {
		stats, err := h.sessionService.GetSessionStats(ctx)
		if err == nil {
			data["metrics"] = gin.H{
				"active_sessions":  stats["generating"],
				"total_sessions":   stats["generating"] + stats["completed"] + stats["expired"],
				"expired_sessions": stats["expired"],
			}
		}
	}

	httpStatus := http.StatusOK
	if data["status"] == "degraded" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response.Response{
		Code:    0,
		Data:    data,
		Message: "health check",
	})
}
