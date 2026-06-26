package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flowtask-server/internal/response"
	"flowtask-server/internal/service"
)

type StudySessionHandler struct {
	sessionService *service.StudySessionService
}

func NewStudySessionHandler(sessionService *service.StudySessionService) *StudySessionHandler {
	return &StudySessionHandler{sessionService: sessionService}
}

func (h *StudySessionHandler) Create(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var input service.CreateSessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	session, err := h.sessionService.Create(userID, input)
	if err != nil {
		response.InternalError(c, "Failed to create session")
		return
	}

	response.Created(c, session)
}

func (h *StudySessionHandler) List(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	sessions, err := h.sessionService.List(userID, nil, nil, nil)
	if err != nil {
		response.InternalError(c, "Failed to list sessions")
		return
	}

	response.Success(c, sessions)
}
