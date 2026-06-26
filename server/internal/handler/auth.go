package handler

import (
	"github.com/gin-gonic/gin"

	"flowtask-server/internal/response"
	"flowtask-server/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input service.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authService.Register(input)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalError(c, "Registration failed")
		return
	}

	response.Created(c, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authService.Login(input)
	if err != nil {
		if err.Error() == "invalid email or password" {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalError(c, "Login failed")
		return
	}

	response.Success(c, result)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(input.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Invalid refresh token")
		return
	}

	response.Success(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&input); err == nil && input.RefreshToken != "" {
		// Extract user_id from auth context if available; ignore errors for logout
		if userIDVal, exists := c.Get("user_id"); exists {
			_ = h.authService.Logout(userIDVal.(string), input.RefreshToken)
		}
	}

	response.Success(c, nil)
}
