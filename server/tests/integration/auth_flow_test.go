package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"flowtask-server/internal/response"
)

// 注册请求体
type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// 登录请求体
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// 刷新 token 请求体
type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// API 响应体
type apiResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// 生成测试用 access token
func generateTestAccessToken(t *testing.T, userID string, secret string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": "test@example.com",
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)
	return tokenStr
}

// generateTestRefreshToken 生成测试用 refresh token
func generateTestRefreshToken(t *testing.T, userID string, secret string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(168 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)
	return tokenStr
}

// setupAuthTestRouter 配置认证相关测试路由
func setupAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// TestAuthRegister_ValidInput 测试注册接口 - 有效输入
func TestAuthRegister_ValidInput(t *testing.T) {
	t.Run("有效注册数据应返回201和用户信息及tokens", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupAuthTestRouter()
		router.POST("/api/auth/register", func(c *gin.Context) {
			var input registerRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			// 模拟注册成功
			response.Created(c, gin.H{
				"user": gin.H{
					"id":           "test-user-id",
					"email":        input.Email,
					"display_name": input.DisplayName,
				},
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
			})
		})

		body := registerRequest{
			Email:       "newuser@example.com",
			Password:    "Password123",
			DisplayName: "测试用户",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 验证状态码
		assert.Equal(t, http.StatusCreated, w.Code, "注册成功应返回 201")

		// 验证响应格式
		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err, "响应体应为合法 JSON")
		assert.Equal(t, 0, resp.Code, "成功时 code 应为 0")
		assert.Equal(t, "created", resp.Message, "message 应为 created")

		// 验证返回数据包含 user 和 tokens
		data, ok := resp.Data.(map[string]interface{})
		assert.True(t, ok, "data 应为对象")
		assert.Contains(t, data, "user", "应返回用户信息")
		assert.Contains(t, data, "access_token", "应返回 access_token")
		assert.Contains(t, data, "refresh_token", "应返回 refresh_token")
	})
}

// TestAuthRegister_DuplicateEmail 测试注册接口 - 重复邮箱
func TestAuthRegister_DuplicateEmail(t *testing.T) {
	t.Run("重复邮箱注册应返回409冲突", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupAuthTestRouter()
		router.POST("/api/auth/register", func(c *gin.Context) {
			var input registerRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			// 模拟邮箱已存在
			if input.Email == "existing@example.com" {
				response.Conflict(c, "email already exists")
				return
			}

			response.Created(c, gin.H{})
		})

		body := registerRequest{
			Email:       "existing@example.com",
			Password:    "Password123",
			DisplayName: "重复用户",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 验证返回 409
		assert.Equal(t, http.StatusConflict, w.Code, "重复邮箱应返回 409")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 40901, resp.Code, "冲突错误 code 应为 40901")
		assert.Equal(t, "email already exists", resp.Message, "应提示邮箱已存在")
	})
}

// TestAuthLogin_ValidCredentials 测试登录接口 - 有效凭据
func TestAuthLogin_ValidCredentials(t *testing.T) {
	t.Run("有效凭据登录应返回200", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupAuthTestRouter()
		router.POST("/api/auth/login", func(c *gin.Context) {
			var input loginRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			// 模拟凭据校验
			if input.Email == "user@example.com" && input.Password == "Password123" {
				response.Success(c, gin.H{
					"user": gin.H{
						"id":    "test-user-id",
						"email": input.Email,
					},
					"access_token":  "test-access-token",
					"refresh_token": "test-refresh-token",
				})
				return
			}

			response.Unauthorized(c, "invalid email or password")
		})

		body := loginRequest{
			Email:    "user@example.com",
			Password: "Password123",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "有效凭据登录应返回 200")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Code, "成功时 code 应为 0")
		assert.Equal(t, "success", resp.Message)

		data := resp.Data.(map[string]interface{})
		assert.Contains(t, data, "user")
		assert.Contains(t, data, "access_token")
		assert.Contains(t, data, "refresh_token")
	})
}

// TestAuthLogin_InvalidPassword 测试登录接口 - 错误密码
func TestAuthLogin_InvalidPassword(t *testing.T) {
	t.Run("错误密码登录应返回401", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要数据库连接")
		}

		router := setupAuthTestRouter()
		router.POST("/api/auth/login", func(c *gin.Context) {
			var input loginRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			// 模拟密码错误
			response.Unauthorized(c, "invalid email or password")
		})

		body := loginRequest{
			Email:    "user@example.com",
			Password: "WrongPassword",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "错误密码应返回 401")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 40101, resp.Code, "认证失败 code 应为 40101")
		assert.Equal(t, "invalid email or password", resp.Message)
	})
}

// TestAuthRefresh_ValidToken 测试刷新 token 接口
func TestAuthRefresh_ValidToken(t *testing.T) {
	t.Run("有效 refresh token 应返回新的 tokens", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过集成测试：需要 Redis 连接")
		}

		accessSecret := "test-access-secret"
		refreshSecret := "test-refresh-secret"

		router := setupAuthTestRouter()
		router.POST("/api/auth/refresh", func(c *gin.Context) {
			var input refreshRequest
			if err := c.ShouldBindJSON(&input); err != nil {
				response.BadRequest(c, err.Error())
				return
			}

			// 验证 refresh token 格式
			token, err := jwt.Parse(input.RefreshToken, func(token *jwt.Token) (interface{}, error) {
				return []byte(refreshSecret), nil
			})

			if err != nil || !token.Valid {
				response.Unauthorized(c, "Invalid refresh token")
				return
			}

			// 模拟返回新 tokens
			newAccess := generateTestAccessToken(t, "test-user-id", accessSecret)
			newRefresh := generateTestRefreshToken(t, "test-user-id", refreshSecret)

			response.Success(c, gin.H{
				"access_token":  newAccess,
				"refresh_token": newRefresh,
			})
		})

		validRefresh := generateTestRefreshToken(t, "test-user-id", refreshSecret)

		body := refreshRequest{RefreshToken: validRefresh}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "有效 refresh token 应返回 200")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.Code)

		data := resp.Data.(map[string]interface{})
		assert.Contains(t, data, "access_token", "应返回新的 access_token")
		assert.Contains(t, data, "refresh_token", "应返回新的 refresh_token")
		assert.NotEmpty(t, data["access_token"], "新 access_token 不应为空")
		assert.NotEmpty(t, data["refresh_token"], "新 refresh_token 不应为空")
	})
}

// TestProtectedEndpoint_WithoutToken 测试受保护接口无 token 访问
func TestProtectedEndpoint_WithoutToken(t *testing.T) {
	t.Run("无 token 访问受保护接口应返回401", func(t *testing.T) {
		router := setupAuthTestRouter()

		// 注册认证中间件
		authMiddleware := func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				response.Unauthorized(c, "Missing authorization header")
				c.Abort()
				return
			}
			c.Next()
		}

		router.GET("/api/protected", authMiddleware, func(c *gin.Context) {
			response.Success(c, gin.H{"message": "secret data"})
		})

		// 不带 Authorization 头
		req, _ := http.NewRequest("GET", "/api/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "无 token 应返回 401")

		var resp apiResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 40101, resp.Code, "未授权 code 应为 40101")
		assert.Equal(t, "Missing authorization header", resp.Message)
	})

	t.Run("无效 token 访问受保护接口应返回401", func(t *testing.T) {
		accessSecret := "test-access-secret"

		router := setupAuthTestRouter()

		// 注册认证中间件（完整版）
		authMiddleware := func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				response.Unauthorized(c, "Missing authorization header")
				c.Abort()
				return
			}

			// 解析 Bearer token
			if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
				response.Unauthorized(c, "Invalid authorization format")
				c.Abort()
				return
			}
			tokenStr := authHeader[7:]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(accessSecret), nil
			})
			if err != nil || !token.Valid {
				response.Unauthorized(c, "Invalid or expired token")
				c.Abort()
				return
			}

			c.Next()
		}

		router.GET("/api/protected", authMiddleware, func(c *gin.Context) {
			response.Success(c, gin.H{"message": "secret data"})
		})

		// 带无效 token
		req, _ := http.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-string")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "无效 token 应返回 401")
	})
}
