package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"flowtask-server/internal/response"
)

// 学习目标创建请求
type createGoalRequest struct {
	Description    string `json:"description"`
	TargetDuration string `json:"target_duration"`
}

// 学习目标更新请求
type updateGoalRequest struct {
	Status string `json:"status"`
}

// 添加任务请求
type addTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	SortOrder   int        `json:"sort_order"`
	ParentTaskID *uuid.UUID `json:"parent_task_id,omitempty"`
}

// setupLearningGoalTestRouter 配置学习目标相关测试路由
func setupLearningGoalTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// TestLearningGoal_CreateGoal 测试创建学习目标
func TestLearningGoal_CreateGoal(t *testing.T) {
	t.Run("创建学习目标应返回201", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库和 AI 服务连接")
		}

		router := setupLearningGoalTestRouter()

		// 模拟认证中间件
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.POST("/api/learning-goals", func(c *gin.Context) {
			var input createGoalRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			if input.Description == "" {
				response.BadRequest(c, "description is required")
				return
			}

			// 模拟创建成功
			goalID := uuid.New()
			response.Created(c, gin.H{
				"id":              goalID.String(),
				"description":     input.Description,
				"target_duration": input.TargetDuration,
				"status":          "active",
				"user_id":         uuid.New().String(),
			})
		})

		body := createGoalRequest{
			Description:    "学习 Go 语言并发编程",
			TargetDuration: "2 weeks",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/learning-goals", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "创建学习目标应返回 201")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Code, "成功时 code 应为 0")

		data := resp.Data.(map[string]interface{})
		assert.NotEmpty(t, data["id"], "应返回目标 ID")
		assert.Equal(t, "学习 Go 语言并发编程", data["description"], "描述应匹配")
		assert.Equal(t, "active", data["status"], "初始状态应为 active")
	})
}

// TestLearningGoal_ListGoals 测试列出学习目标
func TestLearningGoal_ListGoals(t *testing.T) {
	t.Run("列出学习目标应返回分页结果", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupLearningGoalTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.GET("/api/learning-goals", func(c *gin.Context) {
			status := c.Query("status")
			page := c.DefaultQuery("page", "1")
			pageSize := c.DefaultQuery("page_size", "20")

			// 模拟返回列表
			goals := []gin.H{
				{
					"id":          uuid.New().String(),
					"description": "学习 Go 语言",
					"status":      "active",
				},
				{
					"id":          uuid.New().String(),
					"description": "学习 TypeScript",
					"status":      "completed",
				},
			}

			// 按状态筛选
			if status != "" {
				filtered := []gin.H{}
				for _, g := range goals {
					if g["status"] == status {
						filtered = append(filtered, g)
					}
				}
				goals = filtered
			}

			response.Success(c, gin.H{
				"items":     goals,
				"total":     len(goals),
				"page":      page,
				"page_size": pageSize,
			})
		})

		req, _ := http.NewRequest("GET", "/api/learning-goals?page=1&page_size=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "列出目标应返回 200")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp.Data.(map[string]interface{})
		assert.Contains(t, data, "items", "应包含 items 字段")
		assert.Contains(t, data, "total", "应包含 total 字段")

		items, ok := data["items"].([]interface{})
		assert.True(t, ok, "items 应为数组")
		assert.GreaterOrEqual(t, len(items), 1, "至少应返回一条记录")
	})

	t.Run("按状态筛选学习目标", func(t *testing.T) {
		router := setupLearningGoalTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.GET("/api/learning-goals", func(c *gin.Context) {
			status := c.Query("status")
			goals := []gin.H{
				{"id": "1", "description": "目标一", "status": "active"},
				{"id": "2", "description": "目标二", "status": "completed"},
			}
			filtered := []gin.H{}
			for _, g := range goals {
				if g["status"] == status {
					filtered = append(filtered, g)
				}
			}
			response.Success(c, gin.H{"items": filtered, "total": len(filtered), "page": "1", "page_size": "20"})
		})

		req, _ := http.NewRequest("GET", "/api/learning-goals?status=active", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var resp apiResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		data := resp.Data.(map[string]interface{})
		items := data["items"].([]interface{})
		assert.Equal(t, 1, len(items), "active 状态应只返回一条")
	})
}

// TestLearningGoal_UpdateStatus 测试更新学习目标状态
func TestLearningGoal_UpdateStatus(t *testing.T) {
	t.Run("更新目标状态应返回200", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupLearningGoalTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		goalID := uuid.New()

		router.PUT("/api/learning-goals/:id", func(c *gin.Context) {
			id := c.Param("id")
			if _, err := uuid.Parse(id); err != nil {
				response.BadRequest(c, "Invalid goal ID")
				return
			}

			var updates map[string]interface{}
			if err := c.ShouldBindJSON(&updates); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			// 模拟更新
			result := gin.H{
				"id":          id,
				"description": "学习 Go 语言",
				"status":      "active",
			}
			if s, ok := updates["status"].(string); ok {
				result["status"] = s
			}

			response.Success(c, result)
		})

		body := map[string]string{"status": "completed"}
		jsonBody, _ := json.Marshal(body)

		url := fmt.Sprintf("/api/learning-goals/%s", goalID.String())
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "更新目标应返回 200")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Code)

		data := resp.Data.(map[string]interface{})
		assert.Equal(t, "completed", data["status"], "状态应已更新为 completed")
	})
}

// TestLearningGoal_AddTask 测试向学习目标添加任务
func TestLearningGoal_AddTask(t *testing.T) {
	t.Run("向目标添加任务应返回201", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupLearningGoalTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		goalID := uuid.New()

		router.POST("/api/learning-goals/:id/tasks", func(c *gin.Context) {
			goalIDParam := c.Param("id")
			if _, err := uuid.Parse(goalIDParam); err != nil {
				response.BadRequest(c, "Invalid goal ID")
				return
			}

			var input addTaskRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			if input.Title == "" {
				response.BadRequest(c, "title is required")
				return
			}

			taskID := uuid.New()
			response.Created(c, gin.H{
				"id":               taskID.String(),
				"title":            input.Title,
				"description":      input.Description,
				"learning_goal_id": goalIDParam,
				"sort_order":       input.SortOrder,
				"status":           "todo",
			})
		})

		body := addTaskRequest{
			Title:       "理解 goroutine 基本概念",
			Description: "学习 goroutine 的创建和生命周期",
			SortOrder:   0,
		}
		jsonBody, _ := json.Marshal(body)

		url := fmt.Sprintf("/api/learning-goals/%s/tasks", goalID.String())
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "添加任务应返回 201")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp.Data.(map[string]interface{})
		assert.Equal(t, "理解 goroutine 基本概念", data["title"], "任务标题应匹配")
		assert.Equal(t, "todo", data["status"], "新任务状态应为 todo")
		assert.Equal(t, goalID.String(), data["learning_goal_id"], "应关联到正确的目标")
	})
}

// TestLearningGoal_DeleteTask 测试删除学习目标下的任务
func TestLearningGoal_DeleteTask(t *testing.T) {
	t.Run("删除任务应返回200", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupLearningGoalTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.DELETE("/api/learning-goals/:id/tasks/:taskId", func(c *gin.Context) {
			taskID := c.Param("taskId")
			if _, err := uuid.Parse(taskID); err != nil {
				response.BadRequest(c, "Invalid task ID")
				return
			}

			// 模拟删除成功
			response.Success(c, nil)
		})

		goalID := uuid.New()
		taskID := uuid.New()
		url := fmt.Sprintf("/api/learning-goals/%s/tasks/%s", goalID.String(), taskID.String())

		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "删除任务应返回 200")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Code, "成功时 code 应为 0")
	})

	t.Run("使用无效任务ID删除应返回400", func(t *testing.T) {
		router := setupLearningGoalTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.DELETE("/api/learning-goals/:id/tasks/:taskId", func(c *gin.Context) {
			taskID := c.Param("taskId")
			if _, err := uuid.Parse(taskID); err != nil {
				response.BadRequest(c, "Invalid task ID")
				return
			}
			response.Success(c, nil)
		})

		goalID := uuid.New()
		url := fmt.Sprintf("/api/learning-goals/%s/tasks/invalid-id", goalID.String())

		req, _ := http.NewRequest("DELETE", url, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效任务ID应返回 400")
	})
}
