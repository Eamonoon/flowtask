package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flowtask-server/internal/repository"
	"flowtask-server/internal/response"
	"flowtask-server/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) Create(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var input service.CreateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.taskService.Create(userID, input)
	if err != nil {
		response.InternalError(c, "Failed to create task")
		return
	}

	response.Created(c, task)
}

func (h *TaskHandler) List(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	opts := repository.TaskListOptions{
		Status:    c.Query("status"),
		Priority:  c.Query("priority"),
		Search:    c.Query("search"),
		Cursor:    c.Query("cursor"),
		Limit:     limit,
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	// Multi-value status
	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		opts.Statuses = statuses
		opts.Status = "" // clear single to avoid conflict
	}
	// Multi-value priority
	if priorities := c.QueryArray("priority"); len(priorities) > 0 {
		opts.Priorities = priorities
		opts.Priority = ""
	}

	if goalIDStr := c.Query("learning_goal_id"); goalIDStr != "" {
		goalID, _ := uuid.Parse(goalIDStr)
		opts.LearningGoalID = &goalID
	}

	// Label IDs
	if labelIDs := c.QueryArray("label_ids"); len(labelIDs) > 0 {
		for _, idStr := range labelIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				opts.LabelIDs = append(opts.LabelIDs, id)
			}
		}
	}

	// Deadline range
	if from := c.Query("deadline_from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			opts.DeadlineFrom = &t
		}
	}
	if to := c.Query("deadline_to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			opts.DeadlineTo = &endOfDay
		}
	}

	tasks, nextCursor, err := h.taskService.List(userID, opts)
	if err != nil {
		response.InternalError(c, "Failed to list tasks")
		return
	}

	response.Success(c, gin.H{
		"items":       tasks,
		"next_cursor": nextCursor,
		"has_more":    nextCursor != "",
	})
}

func (h *TaskHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	task, err := h.taskService.GetByID(id)
	if err != nil {
		response.NotFound(c, "Task not found")
		return
	}

	response.Success(c, task)
}

func (h *TaskHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	var input service.UpdateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.taskService.Update(id, input)
	if err != nil {
		if err.Error() != "" && len(err.Error()) > 20 {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, "Failed to update task")
		return
	}

	response.Success(c, task)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	deleteSubtasks := c.DefaultQuery("delete_subtasks", "true") == "true"

	if err := h.taskService.Delete(id, deleteSubtasks); err != nil {
		response.InternalError(c, "Failed to delete task")
		return
	}

	response.Success(c, nil)
}

func (h *TaskHandler) AddDependency(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	var input struct {
		DependsOnTaskID uuid.UUID `json:"depends_on_task_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.taskService.AddDependency(taskID, input.DependsOnTaskID); err != nil {
		response.Conflict(c, err.Error())
		return
	}

	response.Created(c, gin.H{
		"task_id":           taskID.String(),
		"depends_on_task_id": input.DependsOnTaskID.String(),
	})
}

func (h *TaskHandler) ListSubtasks(c *gin.Context) {
	parentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	subtasks, err := h.taskService.GetSubtasks(parentID)
	if err != nil {
		response.InternalError(c, "Failed to list subtasks")
		return
	}

	result := make([]gin.H, len(subtasks))
	for i, t := range subtasks {
		completed := t.Status == "done"
		result[i] = gin.H{
			"id":           t.ID.String(),
			"title":        t.Title,
			"is_completed": completed,
			"sort_order":   t.SortOrder,
		}
	}
	response.Success(c, result)
}

func (h *TaskHandler) CreateSubtask(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	parentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	var input struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.taskService.Create(userID, service.CreateTaskInput{
		Title:        input.Title,
		ParentTaskID: &parentID,
	})
	if err != nil {
		response.InternalError(c, "Failed to create subtask")
		return
	}

	response.Created(c, gin.H{
		"id":           task.ID.String(),
		"title":        task.Title,
		"is_completed": false,
		"sort_order":   task.SortOrder,
	})
}

func (h *TaskHandler) UpdateSubtask(c *gin.Context) {
	subtaskID, err := uuid.Parse(c.Param("subtaskId"))
	if err != nil {
		response.BadRequest(c, "Invalid subtask ID")
		return
	}

	var input struct {
		IsCompleted *bool `json:"is_completed"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updateInput := service.UpdateTaskInput{}
	if input.IsCompleted != nil {
		if *input.IsCompleted {
			status := "done"
			updateInput.Status = &status
		} else {
			status := "todo"
			updateInput.Status = &status
		}
	}

	task, err := h.taskService.Update(subtaskID, updateInput)
	if err != nil {
		response.InternalError(c, "Failed to update subtask")
		return
	}

	response.Success(c, gin.H{
		"id":           task.ID.String(),
		"title":        task.Title,
		"is_completed": task.Status == "done",
		"sort_order":   task.SortOrder,
	})
}

func (h *TaskHandler) DeleteSubtask(c *gin.Context) {
	subtaskID, err := uuid.Parse(c.Param("subtaskId"))
	if err != nil {
		response.BadRequest(c, "Invalid subtask ID")
		return
	}

	if err := h.taskService.Delete(subtaskID, false); err != nil {
		response.InternalError(c, "Failed to delete subtask")
		return
	}

	response.Success(c, nil)
}
