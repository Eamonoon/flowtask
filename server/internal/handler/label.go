package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flowtask-server/internal/response"
	"flowtask-server/internal/service"
)

type LabelHandler struct {
	labelService *service.LabelService
}

func NewLabelHandler(labelService *service.LabelService) *LabelHandler {
	return &LabelHandler{labelService: labelService}
}

func (h *LabelHandler) Create(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var input struct {
		Name  string `json:"name" binding:"required,min=1,max=50"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if input.Color == "" {
		input.Color = "#6366f1"
	}

	label, err := h.labelService.Create(userID, input.Name, input.Color)
	if err != nil {
		response.InternalError(c, "Failed to create label")
		return
	}

	response.Created(c, label)
}

func (h *LabelHandler) List(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	labels, err := h.labelService.List(userID)
	if err != nil {
		response.InternalError(c, "Failed to list labels")
		return
	}

	response.Success(c, labels)
}

func (h *LabelHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid label ID")
		return
	}

	if err := h.labelService.Delete(id); err != nil {
		response.InternalError(c, "Failed to delete label")
		return
	}

	response.Success(c, nil)
}
