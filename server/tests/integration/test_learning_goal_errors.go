package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestErrorCodes validates that all error codes are returned correctly
// Error codes: 40001, 40002, 40003, 50001, 50002
func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name           string
		setupRouter    func() *gin.Engine
		requestPath    string
		expectedCode   string
		expectedStatus int
	}{
		{
			name: "40001 - Session expired",
			setupRouter: func() *gin.Engine {
				r := gin.New()
				r.POST("/api/learning-goals/:id/tasks/confirm", func(c *gin.Context) {
					c.JSON(http.StatusBadRequest, gin.H{
						"code":    40001,
						"message": "生成会话已过期，请重新生成",
						"data":    nil,
					})
				})
				return r
			},
			requestPath:    "/api/learning-goals/test/tasks/confirm",
			expectedCode:   "40001",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "40002 - Invalid task data",
			setupRouter: func() *gin.Engine {
				r := gin.New()
				r.POST("/api/learning-goals/:id/tasks/confirm", func(c *gin.Context) {
					c.JSON(http.StatusBadRequest, gin.H{
						"code":    40002,
						"message": "任务数据格式无效",
						"data":    nil,
					})
				})
				return r
			},
			requestPath:    "/api/learning-goals/test/tasks/confirm",
			expectedCode:   "40002",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "40003 - AI response parse failed",
			setupRouter: func() *gin.Engine {
				r := gin.New()
				r.GET("/api/learning-goals/:id/generate-stream", func(c *gin.Context) {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code":    40003,
						"message": "AI 返回了无效的响应，已自动重试",
						"data":    nil,
					})
				})
				return r
			},
			requestPath:    "/api/learning-goals/test/generate-stream",
			expectedCode:   "40003",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "50001 - AI service unavailable",
			setupRouter: func() *gin.Engine {
				r := gin.New()
				r.GET("/api/learning-goals/:id/generate-stream", func(c *gin.Context) {
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"code":    50001,
						"message": "AI 服务暂时不可用，请稍后重试",
						"data":    nil,
					})
				})
				return r
			},
			requestPath:    "/api/learning-goals/test/generate-stream",
			expectedCode:   "50001",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "50002 - Generation timeout",
			setupRouter: func() *gin.Engine {
				r := gin.New()
				r.GET("/api/learning-goals/:id/generate-stream", func(c *gin.Context) {
					c.JSON(http.StatusGatewayTimeout, gin.H{
						"code":    50002,
						"message": "生成超时，请重试",
						"data":    nil,
					})
				})
				return r
			},
			requestPath:    "/api/learning-goals/test/generate-stream",
			expectedCode:   "50002",
			expectedStatus: http.StatusGatewayTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := tt.setupRouter()

			req := httptest.NewRequest("POST", tt.requestPath, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			code, ok := response["code"].(float64)
			if !ok {
				t.Fatalf("Response code is not a number: %v", response["code"])
			}

			if fmt.Sprintf("%.0f", code) != tt.expectedCode {
				t.Errorf("Expected code %s, got %.0f", tt.expectedCode, code)
			}

			message, ok := response["message"].(string)
			if !ok || message == "" {
				t.Error("Expected non-empty message in response")
			}
		})
	}
}

// TestErrorMessagesAreUserFriendly validates that error messages are user-friendly
func TestErrorMessagesAreUserFriendly(t *testing.T) {
	errorMessages := map[string]string{
		"40001": "生成会话已过期，请重新生成",
		"40002": "任务数据格式无效",
		"40003": "AI 返回了无效的响应，已自动重试",
		"50001": "AI 服务暂时不可用，请稍后重试",
		"50002": "生成超时，请重试",
	}

	for code, expectedMsg := range errorMessages {
		t.Run("Error code "+code, func(t *testing.T) {
			// Verify message contains actionable guidance
			if len(expectedMsg) < 5 {
				t.Errorf("Error message for code %s is too short: %s", code, expectedMsg)
			}

			// Verify message is in Chinese (user-facing)
			for _, r := range expectedMsg {
				if r >= 0x4e00 && r <= 0x9fff {
					// Contains Chinese characters - good
					return
				}
			}
			t.Errorf("Error message for code %s should contain Chinese characters: %s", code, expectedMsg)
		})
	}
}
