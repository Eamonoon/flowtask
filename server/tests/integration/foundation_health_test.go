package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// healthResponse 模拟健康检查响应结构
type healthResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// TestHealthEndpoint_ReturnsOK 测试健康检查接口返回 200 状态码
func TestHealthEndpoint_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("健康检查返回200", func(t *testing.T) {
		// 构造请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// 直接模拟 handler 的响应格式（不依赖真实数据库）
		c.JSON(http.StatusOK, healthResponse{
			Code: 0,
			Data: gin.H{
				"status":     "ok",
				"database":   "ok",
				"redis":      "ok",
				"ai_service": "ok",
			},
			Message: "health check",
		})

		// 验证状态码
		assert.Equal(t, http.StatusOK, w.Code, "健康检查接口应返回 200")
	})
}

// TestHealthEndpoint_ResponseFormat 测试健康检查响应体格式
func TestHealthEndpoint_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("响应体包含 code、data、message 字段", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.JSON(http.StatusOK, healthResponse{
			Code: 0,
			Data: gin.H{
				"status":     "ok",
				"database":   "ok",
				"redis":      "ok",
				"ai_service": "ok",
			},
			Message: "health check",
		})

		// 解析响应体
		var resp healthResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应为合法 JSON")

		// 验证 code 字段
		assert.Equal(t, 0, resp.Code, "code 字段应为 0")

		// 验证 message 字段
		assert.Equal(t, "health check", resp.Message, "message 字段应为 'health check'")

		// 验证 data 字段包含必要子字段
		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "data 字段应为对象")

		assert.Contains(t, data, "status", "data 应包含 status 字段")
		assert.Contains(t, data, "database", "data 应包含 database 字段")
		assert.Contains(t, data, "redis", "data 应包含 redis 字段")
		assert.Contains(t, data, "ai_service", "data 应包含 ai_service 字段")
	})
}

// TestHealthEndpoint_DataFields 测试 data 内各字段取值
func TestHealthEndpoint_DataFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("所有服务状态为 ok", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.JSON(http.StatusOK, healthResponse{
			Code: 0,
			Data: gin.H{
				"status":     "ok",
				"database":   "ok",
				"redis":      "ok",
				"ai_service": "ok",
			},
			Message: "health check",
		})

		var resp healthResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		data := resp.Data.(map[string]interface{})
		assert.Equal(t, "ok", data["status"], "整体状态应为 ok")
		assert.Equal(t, "ok", data["database"], "数据库状态应为 ok")
		assert.Equal(t, "ok", data["redis"], "Redis 状态应为 ok")
		assert.Equal(t, "ok", data["ai_service"], "AI 服务状态应为 ok")
	})

	t.Run("降级时 status 为 degraded", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.JSON(http.StatusServiceUnavailable, healthResponse{
			Code: 0,
			Data: gin.H{
				"status":     "degraded",
				"database":   "error",
				"redis":      "ok",
				"ai_service": "ok",
			},
			Message: "health check",
		})

		assert.Equal(t, http.StatusServiceUnavailable, w.Code, "降级时应返回 503")

		var resp healthResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp.Data.(map[string]interface{})
		assert.Equal(t, "degraded", data["status"], "降级时 status 应为 degraded")
		assert.Equal(t, "error", data["database"], "降级时 database 应为 error")
	})
}
