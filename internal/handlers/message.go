package handlers

import (
	"net/http"

	"discord-backend/internal/middleware"
	"discord-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct{}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

func (h *MessageHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	var req struct {
		Content   string     `json:"content" binding:"required"`
		ReplyToID *uuid.UUID `json:"reply_to_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messageService := services.NewMessageService()
	message, err := messageService.Create(channelID, userID, req.Content, req.ReplyToID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

func (h *MessageHandler) Get(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := parseInt(o); err == nil {
			offset = parsed
		}
	}

	messageService := services.NewMessageService()
	messages, err := messageService.GetByChannelID(channelID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	// 获取消息总数
	totalCount, err := messageService.GetCountByChannelID(channelID)
	if err != nil {
		totalCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    totalCount,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *MessageHandler) Delete(c *gin.Context) {
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	messageService := services.NewMessageService()
	if err := messageService.Delete(messageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}
