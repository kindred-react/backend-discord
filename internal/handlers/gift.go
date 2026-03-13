package handlers

import (
	"net/http"

	"discord-backend/internal/middleware"
	"discord-backend/internal/services"
	"discord-backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GiftHandler struct {
	hub *websocket.Hub
}

func NewGiftHandler(hub *websocket.Hub) *GiftHandler {
	return &GiftHandler{
		hub: hub,
	}
}

// SendGift 发送礼物
func (h *GiftHandler) SendGift(c *gin.Context) {
	userID := middleware.GetUserID(c)
	
	var req struct {
		ChannelID string `json:"channel_id" binding:"required"`
		GiftID    string `json:"gift_id" binding:"required"`
		GiftName  string `json:"gift_name" binding:"required"`
		GiftEmoji string `json:"gift_emoji" binding:"required"`
		GiftPrice int    `json:"gift_price" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channelID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// 创建礼物消息
	// content 存储 emoji，voice_url 存储礼物名称
	messageService := services.NewMessageService()
	message, err := messageService.CreateWithType(channelID, userID, req.GiftEmoji, "gift", &req.GiftName, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send gift"})
		return
	}

	// 通过 WebSocket 广播消息
	channelService := services.NewChannelService()
	channel, err := channelService.GetByID(channelID)
	if err == nil {
		h.hub.Broadcast(&websocket.BroadcastMessage{
			Event:     "message:new",
			ChannelID: &channelID,
			GuildID:   channel.GuildID,
			Data:      message,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"gift": gin.H{
			"id":    req.GiftID,
			"name":  req.GiftName,
			"emoji": req.GiftEmoji,
			"price": req.GiftPrice,
		},
	})
}
