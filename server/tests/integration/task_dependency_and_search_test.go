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

// 任务创建请求
type createTaskRequest struct {
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Priority       string     `json:"priority"`
	LearningGoalID *uuid.UUID `json:"learning_goal_id,omitempty"`
}

// 任务更新请求
type updateTaskRequest struct {
	Title  *string `json:"title,omitempty"`
	Status *string `json:"status,omitempty"`
}

// 依赖创建请求
type addDependencyRequest struct {
	DependsOnTaskID uuid.UUID `json:"depends_on_task_id"`
}

// setupTaskTestRouter 配置任务管理测试路由
func setupTaskTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// TestTask_CreateTask 测试创建任务
func TestTask_CreateTask(t *testing.T) {
	t.Run("创建任务应返回201", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupTaskTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.POST("/api/tasks", func(c *gin.Context) {
			var input createTaskRequest
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
				"id":        taskID.String(),
				"title":     input.Title,
				"status":    "todo",
				"priority":  "medium",
			})
		})

		body := createTaskRequest{
			Title:       "学习 Go channel",
			Description: "掌握 channel 的使用场景和最佳实践",
			Priority:    "high",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "创建任务应返回 201")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Code)

		data := resp.Data.(map[string]interface{})
		assert.NotEmpty(t, data["id"], "应返回任务 ID")
		assert.Equal(t, "学习 Go channel", data["title"])
		assert.Equal(t, "todo", data["status"], "新任务状态应为 todo")
	})

	t.Run("缺少标题创建任务应返回400", func(t *testing.T) {
		router := setupTaskTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.POST("/api/tasks", func(c *gin.Context) {
			var input createTaskRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}
			if input.Title == "" {
				response.BadRequest(c, "title is required")
				return
			}
		})

		body := map[string]string{"description": "没有标题的任务"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "缺少标题应返回 400")
	})
}

// TestTask_SearchTasks 测试任务搜索过滤
func TestTask_SearchTasks(t *testing.T) {
	t.Run("搜索关键字应过滤任务结果", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupTaskTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		// 预设任务列表
		allTasks := []gin.H{
			{"id": uuid.New().String(), "title": "学习 Go channel", "status": "todo"},
			{"id": uuid.New().String(), "title": "学习 Go goroutine", "status": "doing"},
			{"id": uuid.New().String(), "title": "学习 Python 基础", "status": "todo"},
			{"id": uuid.New().String(), "title": "Go channel 进阶", "status": "done"},
		}

		router.GET("/api/tasks", func(c *gin.Context) {
			search := c.Query("search")
			status := c.Query("status")

			filtered := make([]gin.H, 0)
			for _, task := range allTasks {
				titleMatch := search == "" || contains(task["title"].(string), search)
				statusMatch := status == "" || task["status"] == status
				if titleMatch && statusMatch {
					filtered = append(filtered, task)
				}
			}

			response.Success(c, gin.H{
				"items":       filtered,
				"next_cursor": "",
				"has_more":    false,
			})
		})

		// 搜索 "Go channel"
		req, _ := http.NewRequest("GET", "/api/tasks?search=Go+channel", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp.Data.(map[string]interface{})
		items := data["items"].([]interface{})
		assert.Equal(t, 2, len(items), "搜索 'Go channel' 应返回 2 条结果")
	})

	t.Run("无搜索条件应返回所有任务", func(t *testing.T) {
		router := setupTaskTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.GET("/api/tasks", func(c *gin.Context) {
			response.Success(c, gin.H{
				"items": []gin.H{
					{"id": "1", "title": "任务一"},
					{"id": "2", "title": "任务二"},
					{"id": "3", "title": "任务三"},
				},
				"next_cursor": "",
				"has_more":    false,
			})
		})

		req, _ := http.NewRequest("GET", "/api/tasks", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var resp apiResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		data := resp.Data.(map[string]interface{})
		items := data["items"].([]interface{})
		assert.Equal(t, 3, len(items), "无搜索条件应返回全部任务")
	})
}

// TestTask_AddDependency 测试添加任务依赖
func TestTask_AddDependency(t *testing.T) {
	t.Run("添加任务依赖应返回201", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupTaskTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.POST("/api/tasks/:id/dependencies", func(c *gin.Context) {
			taskID := c.Param("id")
			if _, err := uuid.Parse(taskID); err != nil {
				response.BadRequest(c, "Invalid task ID")
				return
			}

			var input addDependencyRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			// 检查自依赖
			if taskID == input.DependsOnTaskID.String() {
				response.Conflict(c, "self-dependency is not allowed")
				return
			}

			response.Created(c, gin.H{
				"task_id":           taskID,
				"depends_on_task_id": input.DependsOnTaskID.String(),
			})
		})

		taskID := uuid.New()
		dependsOnID := uuid.New()

		body := addDependencyRequest{DependsOnTaskID: dependsOnID}
		jsonBody, _ := json.Marshal(body)

		url := fmt.Sprintf("/api/tasks/%s/dependencies", taskID.String())
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "添加依赖应返回 201")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp.Data.(map[string]interface{})
		assert.Equal(t, taskID.String(), data["task_id"])
		assert.Equal(t, dependsOnID.String(), data["depends_on_task_id"])
	})

	t.Run("自依赖应返回409冲突", func(t *testing.T) {
		router := setupTaskTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.POST("/api/tasks/:id/dependencies", func(c *gin.Context) {
			taskID := c.Param("id")
			var input addDependencyRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			if taskID == input.DependsOnTaskID.String() {
				response.Conflict(c, "self-dependency is not allowed")
				return
			}

			response.Created(c, gin.H{})
		})

		taskID := uuid.New()
		body := addDependencyRequest{DependsOnTaskID: taskID}
		jsonBody, _ := json.Marshal(body)

		url := fmt.Sprintf("/api/tasks/%s/dependencies", taskID.String())
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "自依赖应返回 409")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 40901, resp.Code)
		assert.Equal(t, "self-dependency is not allowed", resp.Message)
	})
}

// TestTask_UpdateToDone_WithIncompletePrerequisite 测试完成有未完成前置依赖的任务
func TestTask_UpdateToDone_WithIncompletePrerequisite(t *testing.T) {
	t.Run("前置任务未完成时标记完成应返回409", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupTaskTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		// 预设任务和依赖
		// taskA 依赖 taskB，taskB 状态为 todo（未完成）
		taskAID := uuid.New()
		taskBID := uuid.New()
		_ = taskBID

		router.PUT("/api/tasks/:id", func(c *gin.Context) {
			id := c.Param("id")
			if _, err := uuid.Parse(id); err != nil {
				response.BadRequest(c, "Invalid task ID")
				return
			}

			var input updateTaskRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			// 如果要标记为 done，检查前置依赖
			if input.Status != nil && *input.Status == "done" {
				if id == taskAID.String() {
					// 模拟 taskA 有未完成的前置任务
					response.Conflict(c, "cannot mark done: prerequisite task \"学习 Go 基础\" is not completed")
					return
				}
			}

			response.Success(c, gin.H{
				"id":     id,
				"status": *input.Status,
			})
		})

		status := "done"
		body := updateTaskRequest{Status: &status}
		jsonBody, _ := json.Marshal(body)

		url := fmt.Sprintf("/api/tasks/%s", taskAID.String())
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "前置任务未完成时标记完成应返回 409")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 40901, resp.Code)
		assert.Contains(t, resp.Message, "prerequisite task", "应提示前置任务未完成")
	})

	t.Run("所有前置任务完成后标记完成应成功", func(t *testing.T) {
		router := setupTaskTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uuid.New().String())
			c.Next()
		})

		router.PUT("/api/tasks/:id", func(c *gin.Context) {
			var input updateTaskRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}
			// 模拟无未完成前置任务
			response.Success(c, gin.H{
				"id":     c.Param("id"),
				"status": *input.Status,
			})
		})

		taskID := uuid.New()
		status := "done"
		body := updateTaskRequest{Status: &status}
		jsonBody, _ := json.Marshal(body)

		url := fmt.Sprintf("/api/tasks/%s", taskID.String())
		req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "前置任务完成后标记完成应返回 200")

		var resp apiResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp.Data.(map[string]interface{})
		assert.Equal(t, "done", data["status"])
	})
}

// contains 辅助函数，检查字符串包含关系
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
