package service

import (
	"fmt"

	"github.com/google/uuid"

	"flowtask-server/internal/model"
	"flowtask-server/internal/repository"
)

type DashboardCacheInvalidator interface {
	InvalidateCache(userID uuid.UUID)
}

type TaskService struct {
	taskRepo   *repository.TaskRepository
	dashCache  DashboardCacheInvalidator
}

func NewTaskService(taskRepo *repository.TaskRepository, dashCache DashboardCacheInvalidator) *TaskService {
	return &TaskService{taskRepo: taskRepo, dashCache: dashCache}
}

type CreateTaskInput struct {
	Title          string     `json:"title" binding:"required,min=1,max=200"`
	Description    string     `json:"description"`
	Priority       string     `json:"priority"`
	Deadline       *string    `json:"deadline"`
	LearningGoalID *uuid.UUID `json:"learning_goal_id"`
	ParentTaskID   *uuid.UUID `json:"parent_task_id"`
	LabelIDs       []uuid.UUID `json:"label_ids"`
}

type UpdateTaskInput struct {
	Title    *string `json:"title"`
	Status   *string `json:"status"`
	Priority *string `json:"priority"`
	Deadline *string `json:"deadline"`
}

func (s *TaskService) Create(userID uuid.UUID, input CreateTaskInput) (*model.Task, error) {
	task := &model.Task{
		UserID:         userID,
		Title:          input.Title,
		LearningGoalID: input.LearningGoalID,
		ParentTaskID:   input.ParentTaskID,
	}
	if input.Description != "" {
		task.Description = &input.Description
	}
	if input.Priority != "" {
		task.Priority = input.Priority
	}

	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	if s.dashCache != nil {
		s.dashCache.InvalidateCache(userID)
	}
	return task, nil
}

func (s *TaskService) GetByID(id uuid.UUID) (*model.Task, error) {
	return s.taskRepo.FindByID(id)
}

func (s *TaskService) Update(id uuid.UUID, input UpdateTaskInput) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Status != nil {
		if *input.Status == "done" {
			deps, _ := s.taskRepo.GetDependencies(id)
			for _, dep := range deps {
				depTask, err := s.taskRepo.FindByID(dep.DependsOnTaskID)
				if err == nil && depTask.Status != "done" {
					return nil, fmt.Errorf("cannot mark done: prerequisite task %q is not completed", depTask.Title)
				}
			}
		}
		task.Status = *input.Status
	}
	if input.Priority != nil {
		task.Priority = *input.Priority
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, err
	}
	if s.dashCache != nil {
		s.dashCache.InvalidateCache(task.UserID)
	}
	return task, nil
}

func (s *TaskService) Delete(id uuid.UUID, deleteSubtasks bool) error {
	// Fetch task to get user_id before deleting
	task, err := s.taskRepo.FindByID(id)
	if err != nil {
		return err
	}
	userID := task.UserID

	if deleteSubtasks {
		err = s.taskRepo.DeleteWithSubtasks(id)
	} else {
		err = s.taskRepo.Delete(id)
	}
	if err != nil {
		return err
	}
	if s.dashCache != nil {
		s.dashCache.InvalidateCache(userID)
	}
	return nil
}

func (s *TaskService) List(userID uuid.UUID, opts repository.TaskListOptions) ([]model.Task, string, error) {
	return s.taskRepo.ListByUser(userID, opts)
}

func (s *TaskService) AddDependency(taskID, dependsOnID uuid.UUID) error {
	if taskID == dependsOnID {
		return fmt.Errorf("self-dependency is not allowed")
	}
	return s.taskRepo.CreateDependency(taskID, dependsOnID)
}

func (s *TaskService) GetSubtasks(parentID uuid.UUID) ([]model.Task, error) {
	return s.taskRepo.GetSubtasks(parentID)
}
