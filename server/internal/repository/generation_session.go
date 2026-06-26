package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"flowtask-server/internal/model"
)

// GenerationSessionRepository handles database operations for generation sessions
type GenerationSessionRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewGenerationSessionRepository creates a new repository instance
func NewGenerationSessionRepository(db *gorm.DB, redis *redis.Client) *GenerationSessionRepository {
	return &GenerationSessionRepository{
		db:    db,
		redis: redis,
	}
}

// Create creates a new generation session
func (r *GenerationSessionRepository) Create(ctx context.Context, session *model.GenerationSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return err
	}

	// Cache the session
	r.cacheSession(ctx, session)

	return nil
}

// GetByID retrieves a session by ID with Redis caching
func (r *GenerationSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.GenerationSession, error) {
	// Try cache first
	if r.redis != nil {
		cacheKey := fmt.Sprintf("session:%s", id.String())
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var session model.GenerationSession
			if json.Unmarshal([]byte(cached), &session) == nil {
				return &session, nil
			}
		}
	}

	// Fallback to database
	var session model.GenerationSession
	err := r.db.WithContext(ctx).
		Preload("Tasks").
		Where("id = ?", id).
		First(&session).Error
	if err != nil {
		return nil, err
	}

	// Cache the result
	r.cacheSession(ctx, &session)

	return &session, nil
}

// cacheSession caches a session in Redis
func (r *GenerationSessionRepository) cacheSession(ctx context.Context, session *model.GenerationSession) {
	if r.redis == nil {
		return
	}

	cacheKey := fmt.Sprintf("session:%s", session.ID.String())
	data, err := json.Marshal(session)
	if err != nil {
		return
	}

	// Cache for 1 hour
	r.redis.Set(ctx, cacheKey, data, 1*time.Hour)
}

// invalidateCache removes a session from cache
func (r *GenerationSessionRepository) invalidateCache(ctx context.Context, id uuid.UUID) {
	if r.redis == nil {
		return
	}

	cacheKey := fmt.Sprintf("session:%s", id.String())
	r.redis.Del(ctx, cacheKey)
}

// GetByLearningGoalID retrieves active sessions for a learning goal
func (r *GenerationSessionRepository) GetByLearningGoalID(ctx context.Context, learningGoalID uuid.UUID) ([]model.GenerationSession, error) {
	var sessions []model.GenerationSession
	err := r.db.WithContext(ctx).
		Where("learning_goal_id = ? AND status = ?", learningGoalID, "generating").
		Order("created_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// Update updates a generation session
func (r *GenerationSessionRepository) Update(ctx context.Context, session *model.GenerationSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// UpdateStatus updates only the status field
func (r *GenerationSessionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.GenerationSession{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// IncrementTaskCount atomically increments the task count
func (r *GenerationSessionRepository) IncrementTaskCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.GenerationSession{}).
		Where("id = ?", id).
		UpdateColumn("task_count", gorm.Expr("task_count + 1")).Error
}

// Delete deletes a generation session and its generated tasks (cascade)
func (r *GenerationSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.GenerationSession{}).Error
}

// CleanupExpiredSessions marks expired sessions and returns count of affected rows
func (r *GenerationSessionRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.GenerationSession{}).
		Where("expires_at < ? AND status = ?", time.Now(), "generating").
		Update("status", "expired")
	return result.RowsAffected, result.Error
}

// GetExpiredSessions retrieves all expired sessions for cleanup
func (r *GenerationSessionRepository) GetExpiredSessions(ctx context.Context) ([]model.GenerationSession, error) {
	var sessions []model.GenerationSession
	err := r.db.WithContext(ctx).
		Where("expires_at < ? AND status = ?", time.Now(), "expired").
		Find(&sessions).Error
	return sessions, err
}

// CountByStatus counts sessions by status
func (r *GenerationSessionRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Status string
		Count  int64
	}

	err := r.db.WithContext(ctx).
		Model(&model.GenerationSession{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}
