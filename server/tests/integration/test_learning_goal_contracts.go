package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCreateLearningGoal_Contract tests the POST /api/learning-goals endpoint contract
func TestCreateLearningGoal_Contract(t *testing.T) {
	// Setup router with mock handler
	router := gin.New()
	router.POST("/api/learning-goals", func(c *gin.Context) {
		var input struct {
			Description    string `json:"description" binding:"required"`
			TargetDuration string `json:"target_duration"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"code": 0,
			"data": gin.H{
				"session_id":       "test-session-id",
				"learning_goal_id": "test-goal-id",
				"status":           "generating",
			},
			"message": "success",
		})
	})

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedCode   float64
	}{
		{
			name:           "valid request",
			body:           map[string]string{"description": "我想学 Go"},
			expectedStatus: http.StatusCreated,
			expectedCode:   0,
		},
		{
			name:           "with target duration",
			body:           map[string]string{"description": "我想学 Go", "target_duration": "2个月"},
			expectedStatus: http.StatusCreated,
			expectedCode:   0,
		},
		{
			name:           "missing description",
			body:           map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/learning-goals", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.expectedCode, response["code"])

			if tt.expectedStatus == http.StatusCreated {
				data := response["data"].(map[string]interface{})
				assert.NotEmpty(t, data["session_id"])
				assert.NotEmpty(t, data["learning_goal_id"])
				assert.Equal(t, "generating", data["status"])
			}
		})
	}
}

// TestGenerateStream_Contract tests the GET /api/learning-goals/:id/generate-stream endpoint contract
func TestGenerateStream_Contract(t *testing.T) {
	router := gin.New()
	router.GET("/api/learning-goals/:id/generate-stream", func(c *gin.Context) {
		sessionID := c.Query("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "session_id is required",
				"data":    nil,
			})
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		// Send a task event
		task := map[string]interface{}{
			"id":    "task-1",
			"title": "Learn Go",
		}
		data, _ := json.Marshal(task)
		c.Writer.WriteString("event: task\ndata: " + string(data) + "\n\n")
		c.Writer.Flush()

		// Send done event
		c.Writer.WriteString("event: done\ndata: {\"task_count\":1}\n\n")
		c.Writer.Flush()
	})

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
	}{
		{
			name:           "valid request",
			sessionID:      "test-session",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing session_id",
			sessionID:      "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/learning-goals/test-goal/generate-stream"
			if tt.sessionID != "" {
				url += "?session_id=" + tt.sessionID
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, w.Header().Get("Content-Type"), "text/event-stream")
				assert.Contains(t, w.Body.String(), "event: task")
				assert.Contains(t, w.Body.String(), "event: done")
			}
		})
	}
}

// TestConfirmTasks_Contract tests the POST /api/learning-goals/:id/tasks/confirm endpoint contract
func TestConfirmTasks_Contract(t *testing.T) {
	router := gin.New()
	router.POST("/api/learning-goals/:id/tasks/confirm", func(c *gin.Context) {
		var input struct {
			SessionID string `json:"session_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}

		if input.SessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    40001,
				"message": "生成会话已过期，请重新生成",
				"data":    nil,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"learning_goal_id": "test-goal-id",
				"saved_task_count": 5,
				"message":          "学习计划已保存",
			},
			"message": "success",
		})
	})

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		expectedCode   float64
	}{
		{
			name:           "valid request",
			body:           map[string]string{"session_id": "test-session"},
			expectedStatus: http.StatusOK,
			expectedCode:   0,
		},
		{
			name:           "expired session",
			body:           map[string]string{"session_id": ""},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   40001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/learning-goals/test-goal/tasks/confirm", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.expectedCode, response["code"])

			if tt.expectedStatus == http.StatusOK {
				data := response["data"].(map[string]interface{})
				assert.NotEmpty(t, data["learning_goal_id"])
				assert.NotNil(t, data["saved_task_count"])
			}
		})
	}
}
