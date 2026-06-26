package handler

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"flowtask-server/internal/response"
	"flowtask-server/internal/service"
)

type AIChatHandler struct {
	chatService *service.AIChatService
}

func NewAIChatHandler(chatService *service.AIChatService) *AIChatHandler {
	return &AIChatHandler{chatService: chatService}
}

func (h *AIChatHandler) Chat(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	var input service.ChatInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	err := h.chatService.Chat(userID, input,
		func(delta string) error {
			data, _ := json.Marshal(map[string]string{"content": delta})
			_, err := fmt.Fprintf(c.Writer, "event: delta\ndata: %s\n\n", data)
			c.Writer.Flush()
			return err
		},
		func(fullContent string, convID uuid.UUID) error {
			data, _ := json.Marshal(map[string]interface{}{
				"full_content":    fullContent,
				"conversation_id": convID.String(),
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

func (h *AIChatHandler) ListConversations(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	convs, err := h.chatService.ListConversations(userID)
	if err != nil {
		response.InternalError(c, "Failed to list conversations")
		return
	}

	response.Success(c, convs)
}

func (h *AIChatHandler) GetMessages(c *gin.Context) {
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	messages, err := h.chatService.GetMessages(convID)
	if err != nil {
		response.InternalError(c, "Failed to get messages")
		return
	}

	response.Success(c, messages)
}

func (h *AIChatHandler) DeleteConversation(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid conversation ID")
		return
	}

	if err := h.chatService.DeleteConversation(userID, convID); err != nil {
		response.InternalError(c, "Failed to delete conversation")
		return
	}

	response.Success(c, nil)
}
