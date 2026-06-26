package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"flowtask-server/internal/model"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) FindByID(id uuid.UUID) (*model.Task, error) {
	var task model.Task
	err := r.db.Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

func (r *TaskRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.Task{}).Error
}

func (r *TaskRepository) DeleteWithSubtasks(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("parent_task_id = ?", id).Delete(&model.Task{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&model.Task{}).Error
	})
}

func (r *TaskRepository) ListByGoal(goalID uuid.UUID) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Where("learning_goal_id = ?", goalID).
		Order("sort_order ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) ListByUser(userID uuid.UUID, opts TaskListOptions) ([]model.Task, string, error) {
	var tasks []model.Task
	query := r.db.Where("user_id = ?", userID)

	// Single status (backward compat)
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	// Multi-status filter
	if len(opts.Statuses) > 0 {
		query = query.Where("status IN ?", opts.Statuses)
	}
	// Single priority (backward compat)
	if opts.Priority != "" {
		query = query.Where("priority = ?", opts.Priority)
	}
	// Multi-priority filter
	if len(opts.Priorities) > 0 {
		query = query.Where("priority IN ?", opts.Priorities)
	}
	if opts.LearningGoalID != nil {
		query = query.Where("learning_goal_id = ?", opts.LearningGoalID)
	}
	if opts.Search != "" {
		search := "%" + opts.Search + "%"
		query = query.Where("(title ILIKE ? OR description ILIKE ?)", search, search)
	}
	// Label filter via join
	if len(opts.LabelIDs) > 0 {
		query = query.Joins("INNER JOIN task_labels ON task_labels.task_id = tasks.id").
			Where("task_labels.label_id IN ?", opts.LabelIDs)
	}
	// Deadline range
	if opts.DeadlineFrom != nil {
		query = query.Where("deadline >= ?", opts.DeadlineFrom)
	}
	if opts.DeadlineTo != nil {
		query = query.Where("deadline <= ?", opts.DeadlineTo)
	}
	if opts.Cursor != "" {
		cursorID, err := uuid.Parse(opts.Cursor)
		if err == nil {
			query = query.Where("id < ?", cursorID)
		}
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Sorting
	orderClause := "created_at DESC"
	if opts.SortBy != "" {
		// Whitelist allowed sort columns
		allowedSorts := map[string]bool{
			"created_at": true, "updated_at": true, "deadline": true,
			"priority": true, "sort_order": true,
		}
		if allowedSorts[opts.SortBy] {
			direction := "ASC"
			if opts.SortBy == "created_at" || opts.SortBy == "updated_at" {
				direction = "DESC"
			}
			if opts.SortOrder == "desc" {
				direction = "DESC"
			} else if opts.SortOrder == "asc" {
				direction = "ASC"
			}
			orderClause = opts.SortBy + " " + direction
		}
	}

	err := query.Order(orderClause).Limit(limit + 1).Find(&tasks).Error
	if err != nil {
		return nil, "", err
	}

	nextCursor := ""
	if len(tasks) > limit {
		tasks = tasks[:limit]
		nextCursor = tasks[len(tasks)-1].ID.String()
	}

	return tasks, nextCursor, nil
}

func (r *TaskRepository) CreateDependency(taskID, dependsOnID uuid.UUID) error {
	dep := model.TaskDependency{
		TaskID:          taskID,
		DependsOnTaskID: dependsOnID,
	}
	return r.db.Create(&dep).Error
}

func (r *TaskRepository) GetDependencies(taskID uuid.UUID) ([]model.TaskDependency, error) {
	var deps []model.TaskDependency
	err := r.db.Where("task_id = ?", taskID).Find(&deps).Error
	return deps, err
}

func (r *TaskRepository) GetSubtasks(parentID uuid.UUID) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Where("parent_task_id = ?", parentID).
		Order("sort_order ASC").
		Find(&tasks).Error
	return tasks, err
}

type TaskListOptions struct {
	Status         string
	Statuses       []string
	Priority       string
	Priorities     []string
	LearningGoalID *uuid.UUID
	Search         string
	Cursor         string
	Limit          int
	LabelIDs       []uuid.UUID
	DeadlineFrom   *time.Time
	DeadlineTo     *time.Time
	SortBy         string
	SortOrder      string
}
