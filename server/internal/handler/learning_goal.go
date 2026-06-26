package handler

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flowtask-server/internal/response"
	"flowtask-server/internal/service"
)

type LearningGoalHandler struct {
	goalService *service.LearningGoalService
}

func NewLearningGoalHandler(goalService *service.LearningGoalService) *LearningGoalHandler {
	return &LearningGoalHandler{goalService: goalService}
}

func (h *LearningGoalHandler) Create(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var input service.CreateGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	err := h.goalService.GeneratePlanStream(userID, input,
		func(task map[string]interface{}) error {
			data, _ := json.Marshal(task)
			_, err := fmt.Fprintf(c.Writer, "event: task\ndata: %s\n\n", data)
			c.Writer.Flush()
			return err
		},
		func(goalID uuid.UUID, taskCount int) error {
			data, _ := json.Marshal(map[string]interface{}{
				"learning_goal_id": goalID.String(),
				"task_count":       taskCount,
			})
			_, err := fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", data)
			c.Writer.Flush()
			return err
		},
	)

	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", data)
		c.Writer.Flush()
	}
}

func (h *LearningGoalHandler) List(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	goals, total, err := h.goalService.List(userID, status, page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list goals")
		return
	}

	response.Success(c, gin.H{
		"items":     goals,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *LearningGoalHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid goal ID")
		return
	}

	goal, err := h.goalService.GetByID(id)
	if err != nil {
		response.NotFound(c, "Goal not found")
		return
	}

	response.Success(c, goal)
}

func (h *LearningGoalHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid goal ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	goal, err := h.goalService.Update(id, updates)
	if err != nil {
		response.InternalError(c, "Failed to update goal")
		return
	}

	response.Success(c, goal)
}

func (h *LearningGoalHandler) AddTask(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid goal ID")
		return
	}

	var input struct {
		Title        string     `json:"title" binding:"required"`
		Description  string     `json:"description"`
		ParentTaskID *uuid.UUID `json:"parent_task_id"`
		SortOrder    int        `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := h.goalService.AddTask(goalID, userID, input.Title, input.Description, input.ParentTaskID, input.SortOrder)
	if err != nil {
		response.InternalError(c, "Failed to add task")
		return
	}

	response.Created(c, task)
}

func (h *LearningGoalHandler) DeleteTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("taskId"))
	if err != nil {
		response.BadRequest(c, "Invalid task ID")
		return
	}

	if err := h.goalService.DeleteTask(taskID); err != nil {
		response.InternalError(c, "Failed to delete task")
		return
	}

	response.Success(c, nil)
}

// CreateWithSession creates a learning goal and returns a session_id for streaming
func (h *LearningGoalHandler) CreateWithSession(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var input service.CreateGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create goal and session
	result, err := h.goalService.CreateGoalWithSession(userID, input)
	if err != nil {
		response.InternalError(c, "Failed to create learning goal: "+err.Error())
		return
	}

	response.Created(c, result)
}

// GenerateStream handles SSE streaming of generated tasks
func (h *LearningGoalHandler) GenerateStream(c *gin.Context) {
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid goal ID")
		return
	}

	sessionIDStr := c.Query("session_id")
	if sessionIDStr == "" {
		response.BadRequest(c, "session_id is required")
		return
	}
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid session_id")
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Start streaming
	err = h.goalService.GeneratePlanStreamWithSession(goalID, sessionID,
		func(task map[string]interface{}) error {
			data, _ := json.Marshal(task)
			_, err := fmt.Fprintf(c.Writer, "event: task\ndata: %s\n\n", data)
			c.Writer.Flush()
			return err
		},
		func(taskCount int) error {
			data, _ := json.Marshal(map[string]interface{}{
				"task_count": taskCount,
			})
			_, err := fmt.Fprintf(c.Writer, "event: progress\ndata: %s\n\n", data)
			c.Writer.Flush()
			return err
		},
		func(taskCount int) error {
			data, _ := json.Marshal(map[string]interface{}{
				"learning_goal_id": goalID.String(),
				"task_count":       taskCount,
			})
			_, err := fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", data)
			c.Writer.Flush()
			return err
		},
	)

	if err != nil {
		data, _ := json.Marshal(map[string]string{
			"code":    "STREAM_ERROR",
			"message": err.Error(),
		})
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", data)
		c.Writer.Flush()
	}
}

// ConfirmTasks confirms and saves generated tasks to the main tasks table
func (h *LearningGoalHandler) ConfirmTasks(c *gin.Context) {
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid goal ID")
		return
	}

	var input struct {
		SessionID string `json:"session_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	sessionID, err := uuid.Parse(input.SessionID)
	if err != nil {
		response.BadRequest(c, "Invalid session_id")
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	result, err := h.goalService.ConfirmAndSaveTasks(goalID, sessionID, userID)
	if err != nil {
		response.InternalError(c, "Failed to confirm tasks: "+err.Error())
		return
	}

	response.Success(c, result)
}

// Regenerate creates a new generation session for an existing goal
func (h *LearningGoalHandler) Regenerate(c *gin.Context) {
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid goal ID")
		return
	}

	result, err := h.goalService.RegeneratePlan(goalID)
	if err != nil {
		response.InternalError(c, "Failed to regenerate: "+err.Error())
		return
	}

	response.Success(c, result)
}
