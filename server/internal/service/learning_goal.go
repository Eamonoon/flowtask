package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"flowtask-server/internal/ai"
	"flowtask-server/internal/model"
	"flowtask-server/internal/repository"
)

type LearningGoalService struct {
	goalRepo       *repository.LearningGoalRepository
	taskRepo       *repository.TaskRepository
	aiClient       *ai.Client
	sessionService *GenerationSessionService
}

func NewLearningGoalService(
	goalRepo *repository.LearningGoalRepository,
	taskRepo *repository.TaskRepository,
	aiClient *ai.Client,
	sessionService *GenerationSessionService,
) *LearningGoalService {
	return &LearningGoalService{
		goalRepo:       goalRepo,
		taskRepo:       taskRepo,
		aiClient:       aiClient,
		sessionService: sessionService,
	}
}

type CreateGoalInput struct {
	Description    string `json:"description" binding:"required,min=1,max=2000"`
	TargetDuration string `json:"target_duration"`
}

func (s *LearningGoalService) Create(userID uuid.UUID, input CreateGoalInput) (*model.LearningGoal, error) {
	goal := &model.LearningGoal{
		UserID:      userID,
		Description: input.Description,
		Status:      "active",
	}
	if input.TargetDuration != "" {
		goal.TargetDuration = &input.TargetDuration
	}

	if err := s.goalRepo.Create(goal); err != nil {
		return nil, fmt.Errorf("create goal: %w", err)
	}

	return goal, nil
}

func (s *LearningGoalService) GeneratePlanStream(userID uuid.UUID, input CreateGoalInput, onTask func(task map[string]interface{}) error, onDone func(goalID uuid.UUID, taskCount int) error) error {
	goal, err := s.Create(userID, input)
	if err != nil {
		return err
	}

	prompt := ai.BuildLearningPlanPrompt(input.Description, input.TargetDuration)

	var fullContent string
	err = s.aiClient.ChatStream(prompt, func(delta string) error {
		fullContent += delta
		return nil
	})
	if err != nil {
		return fmt.Errorf("AI stream: %w", err)
	}

	plan, err := ai.ParsePlanFromAI(fullContent)
	if err != nil {
		return fmt.Errorf("parse plan: %w", err)
	}

	taskCount := 0
	taskIDMap := make(map[string]uuid.UUID)

	for i, genTask := range plan.Tasks {
		task := &model.Task{
			UserID:         userID,
			LearningGoalID: &goal.ID,
			Title:          genTask.Title,
			Description:    &genTask.Description,
			EstimatedDuration: &genTask.EstimatedDuration,
			SortOrder:      i,
		}
		if genTask.RecommendedResources != nil {
			task.SetResourcesJSON(genTask.RecommendedResources)
		}

		if err := s.taskRepo.Create(task); err != nil {
			return fmt.Errorf("create task: %w", err)
		}
		taskIDMap[genTask.Title] = task.ID
		taskCount++

		taskData := map[string]interface{}{
			"id":                 task.ID.String(),
			"title":              task.Title,
			"description":        task.Description,
			"estimated_duration": task.EstimatedDuration,
			"recommended_resources": genTask.RecommendedResources,
			"parent_task_id":     nil,
			"sort_order":         task.SortOrder,
		}
		if err := onTask(taskData); err != nil {
			return err
		}

		for j, subTask := range genTask.Subtasks {
			sub := &model.Task{
				UserID:         userID,
				LearningGoalID: &goal.ID,
				ParentTaskID:   &task.ID,
				Title:          subTask.Title,
				Description:    &subTask.Description,
				EstimatedDuration: &subTask.EstimatedDuration,
				SortOrder:      j,
			}
			if err := s.taskRepo.Create(sub); err != nil {
				return fmt.Errorf("create subtask: %w", err)
			}
			taskCount++

			subData := map[string]interface{}{
				"id":                 sub.ID.String(),
				"title":              sub.Title,
				"description":        sub.Description,
				"estimated_duration": sub.EstimatedDuration,
				"recommended_resources": []interface{}{},
				"parent_task_id":     task.ID.String(),
				"sort_order":         sub.SortOrder,
			}
			if err := onTask(subData); err != nil {
				return err
			}
		}
	}

	for _, genTask := range plan.Tasks {
		for _, depTitle := range genTask.Dependencies {
			if depID, ok := taskIDMap[depTitle]; ok {
				if taskID, ok := taskIDMap[genTask.Title]; ok {
					s.taskRepo.CreateDependency(taskID, depID)
				}
			}
		}
	}

	return onDone(goal.ID, taskCount)
}

func (s *LearningGoalService) List(userID uuid.UUID, status string, page, pageSize int) ([]model.LearningGoal, int64, error) {
	return s.goalRepo.ListByUser(userID, status, page, pageSize)
}

func (s *LearningGoalService) GetByID(id uuid.UUID) (*model.LearningGoal, error) {
	return s.goalRepo.FindByID(id)
}

func (s *LearningGoalService) Update(id uuid.UUID, updates map[string]interface{}) (*model.LearningGoal, error) {
	goal, err := s.goalRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if v, ok := updates["status"].(string); ok {
		goal.Status = v
	}
	if v, ok := updates["description"].(string); ok {
		goal.Description = v
	}

	if err := s.goalRepo.Update(goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *LearningGoalService) AddTask(goalID, userID uuid.UUID, title, description string, parentTaskID *uuid.UUID, sortOrder int) (*model.Task, error) {
	task := &model.Task{
		UserID:         userID,
		LearningGoalID: &goalID,
		Title:          title,
		Description:    &description,
		ParentTaskID:   parentTaskID,
		SortOrder:      sortOrder,
	}
	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *LearningGoalService) DeleteTask(taskID uuid.UUID) error {
	return s.taskRepo.DeleteWithSubtasks(taskID)
}

// CreateGoalWithSessionInput is the input for CreateGoalWithSession
type CreateGoalWithSessionInput struct {
	Description    string `json:"description" binding:"required"`
	TargetDuration string `json:"target_duration"`
}

// CreateGoalWithSession creates a learning goal and a generation session
func (s *LearningGoalService) CreateGoalWithSession(userID uuid.UUID, input CreateGoalInput) (map[string]interface{}, error) {
	// Create the learning goal
	goal := &model.LearningGoal{
		UserID:      userID,
		Description: input.Description,
	}
	if input.TargetDuration != "" {
		goal.TargetDuration = &input.TargetDuration
	}
	if err := s.goalRepo.Create(goal); err != nil {
		return nil, fmt.Errorf("failed to create goal: %w", err)
	}

	// Create generation session
	session, err := s.sessionService.CreateSession(context.Background(), goal.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return map[string]interface{}{
		"session_id":       session.ID.String(),
		"learning_goal_id": goal.ID.String(),
		"status":           "generating",
	}, nil
}

// GeneratePlanStreamWithSession generates a plan and streams tasks via SSE
func (s *LearningGoalService) GeneratePlanStreamWithSession(
	goalID uuid.UUID,
	sessionID uuid.UUID,
	onTask func(map[string]interface{}) error,
	onProgress func(int) error,
	onDone func(int) error,
) error {
	// Get the goal
	goal, err := s.goalRepo.FindByID(goalID)
	if err != nil {
		return fmt.Errorf("goal not found: %w", err)
	}

	// Get session to verify it exists and is active
	session, err := s.sessionService.GetSession(context.Background(), sessionID)
	if err != nil {
		return fmt.Errorf("session not found or expired: %w", err)
	}

	if !session.IsGenerating() {
		return fmt.Errorf("session is not in generating state")
	}

	// Build prompt
	prompt := s.buildGoalPrompt(goal)

	// Generate tasks with AI
	tasks, err := s.aiClient.GenerateTasks(prompt)
	if err != nil {
		return fmt.Errorf("failed to generate tasks: %w", err)
	}

	// Stream tasks
	taskCount := 0
	for i, task := range tasks {
		// Create generated task
		genTask := &model.GeneratedTask{
			SessionID:         sessionID,
			Title:             task.Title,
			Description:       task.Description,
			EstimatedDuration: task.EstimatedDuration,
			SortOrder:         i,
		}

		// Convert recommended_resources from AI response
		if task.RecommendedResources != nil {
			genTask.SetResourcesJSON(task.RecommendedResources)
		}

		// Add to session
		if err := s.sessionService.AddTask(context.Background(), sessionID, genTask); err != nil {
			return fmt.Errorf("failed to save task: %w", err)
		}

		// Stream to client
		taskMap := map[string]interface{}{
			"id":                  genTask.ID.String(),
			"title":               genTask.Title,
			"description":         genTask.Description,
			"estimated_duration":  genTask.EstimatedDuration,
			"parent_task_id":      nil,
			"sort_order":          genTask.SortOrder,
		}
		if genTask.ParentTaskID != nil {
			taskMap["parent_task_id"] = genTask.ParentTaskID.String()
		}

		if err := onTask(taskMap); err != nil {
			return fmt.Errorf("failed to stream task: %w", err)
		}

		taskCount++

		// Send progress every 3 tasks
		if taskCount%3 == 0 {
			if err := onProgress(taskCount); err != nil {
				return fmt.Errorf("failed to send progress: %w", err)
			}
		}
	}

	// Send final progress
	if err := onProgress(taskCount); err != nil {
		return fmt.Errorf("failed to send final progress: %w", err)
	}

	return onDone(taskCount)
}

// ConfirmAndSaveTasks confirms and saves generated tasks to the main tasks table
func (s *LearningGoalService) ConfirmAndSaveTasks(goalID uuid.UUID, sessionID uuid.UUID, userID uuid.UUID) (map[string]interface{}, error) {
	// Verify session exists and belongs to the goal
	session, err := s.sessionService.GetSession(context.Background(), sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.LearningGoalID != goalID {
		return nil, fmt.Errorf("session does not belong to this goal")
	}

	// Move tasks to main table
	count, err := s.sessionService.MoveTasksToMainTable(context.Background(), sessionID, goalID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to move tasks: %w", err)
	}

	return map[string]interface{}{
		"learning_goal_id": goalID.String(),
		"saved_task_count": count,
		"message":          "学习计划已保存",
	}, nil
}

// RegeneratePlan creates a new generation session for an existing goal
func (s *LearningGoalService) RegeneratePlan(goalID uuid.UUID) (map[string]interface{}, error) {
	// Verify goal exists
	_, err := s.goalRepo.FindByID(goalID)
	if err != nil {
		return nil, fmt.Errorf("goal not found: %w", err)
	}

	// Clean up any existing active sessions
	existingSession, err := s.sessionService.GetActiveSession(context.Background(), goalID)
	if err == nil && existingSession != nil {
		// Delete old session
		_ = s.sessionService.DeleteSession(context.Background(), existingSession.ID)
	}

	// Create new session
	session, err := s.sessionService.CreateSession(context.Background(), goalID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return map[string]interface{}{
		"session_id":       session.ID.String(),
		"learning_goal_id": goalID.String(),
		"status":           "generating",
	}, nil
}

// buildGoalPrompt builds the AI prompt for generating tasks
func (s *LearningGoalService) buildGoalPrompt(goal *model.LearningGoal) string {
	prompt := "根据以下学习目标，生成一个详细的学习计划任务列表。\n\n"
	prompt += "学习目标：" + goal.Description + "\n"
	if goal.TargetDuration != nil && *goal.TargetDuration != "" {
		prompt += "目标时长：" + *goal.TargetDuration + "\n"
	}
	prompt += "\n请以 JSON 数组格式返回任务列表，每个任务包含以下字段：\n"
	prompt += "- title: 任务标题\n"
	prompt += "- description: 任务描述\n"
	prompt += "- estimated_duration: 预估时长\n"
	prompt += "- recommended_resources: 推荐资源数组\n\n"
	prompt += "示例格式：\n"
	prompt += `[
  {
    "title": "学习基础概念",
    "description": "了解核心概念和基本原理",
    "estimated_duration": "1周",
    "recommended_resources": ["https://example.com/guide"]
  }
]`
	return prompt
}
