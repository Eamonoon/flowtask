package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"flowtask-server/internal/model"
)

// GeneratedTaskRepository handles database operations for generated tasks
type GeneratedTaskRepository struct {
	db *gorm.DB
}

// NewGeneratedTaskRepository creates a new repository instance
func NewGeneratedTaskRepository(db *gorm.DB) *GeneratedTaskRepository {
	return &GeneratedTaskRepository{db: db}
}

// Create creates a new generated task
func (r *GeneratedTaskRepository) Create(ctx context.Context, task *model.GeneratedTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// CreateBatch creates multiple generated tasks in a single transaction
func (r *GeneratedTaskRepository) CreateBatch(ctx context.Context, tasks []model.GeneratedTask) error {
	if len(tasks) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(tasks, 100).Error
}

// GetByID retrieves a generated task by ID
func (r *GeneratedTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.GeneratedTask, error) {
	var task model.GeneratedTask
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetBySessionID retrieves all generated tasks for a session
func (r *GeneratedTaskRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.GeneratedTask, error) {
	var tasks []model.GeneratedTask
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("sort_order ASC, created_at ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetRootTasksBySessionID retrieves root tasks (no parent) for a session
func (r *GeneratedTaskRepository) GetRootTasksBySessionID(ctx context.Context, sessionID uuid.UUID) ([]model.GeneratedTask, error) {
	var tasks []model.GeneratedTask
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND parent_task_id IS NULL", sessionID).
		Order("sort_order ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetSubtasksByParentID retrieves subtasks for a parent task
func (r *GeneratedTaskRepository) GetSubtasksByParentID(ctx context.Context, parentID uuid.UUID) ([]model.GeneratedTask, error) {
	var tasks []model.GeneratedTask
	err := r.db.WithContext(ctx).
		Where("parent_task_id = ?", parentID).
		Order("sort_order ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// DeleteBySessionID deletes all generated tasks for a session
func (r *GeneratedTaskRepository) DeleteBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Delete(&model.GeneratedTask{}).Error
}

// DeleteByID deletes a specific generated task
func (r *GeneratedTaskRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.GeneratedTask{}).Error
}

// CountBySessionID counts tasks in a session
func (r *GeneratedTaskRepository) CountBySessionID(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GeneratedTask{}).
		Where("session_id = ?", sessionID).
		Count(&count).Error
	return count, err
}

// MoveToTasks moves generated tasks to the main tasks table after confirmation
func (r *GeneratedTaskRepository) MoveToTasks(ctx context.Context, sessionID uuid.UUID, learningGoalID uuid.UUID, userID uuid.UUID) (int64, error) {
	// Use raw SQL for atomic operation
	result := r.db.WithContext(ctx).Exec(`
		INSERT INTO tasks (id, user_id, learning_goal_id, title, description, estimated_duration, recommended_resources, parent_task_id, sort_order, status, created_at, updated_at)
		SELECT
			gen_random_uuid(),
			?,
			?,
			title,
			description,
			estimated_duration,
			recommended_resources,
			parent_task_id,
			sort_order,
			'todo',
			NOW(),
			NOW()
		FROM generated_tasks
		WHERE session_id = ?
	`, userID, learningGoalID, sessionID)

	if result.Error != nil {
		return 0, result.Error
	}

	// Delete the generated tasks after moving
	err := r.DeleteBySessionID(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected, nil
}
