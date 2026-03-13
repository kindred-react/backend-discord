package handlers

import (
	"net/http"

	"discord-backend/internal/middleware"
	"discord-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VoiceHandler struct{}

func NewVoiceHandler() *VoiceHandler {
	return &VoiceHandler{}
}

func (h *VoiceHandler) Join(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		ChannelID uuid.UUID `json:"channel_id" binding:"required"`
		GuildID   uuid.UUID `json:"guild_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	voiceService := services.NewVoiceService()
	state, err := voiceService.JoinChannel(userID, req.ChannelID, req.GuildID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join voice channel"})
		return
	}

	c.JSON(http.StatusOK, state)
}

func (h *VoiceHandler) Leave(c *gin.Context) {
	userID := middleware.GetUserID(c)

	voiceService := services.NewVoiceService()
	if err := voiceService.LeaveChannel(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave voice channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Left voice channel"})
}

func (h *VoiceHandler) GetChannelState(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	voiceService := services.NewVoiceService()
	states, err := voiceService.GetByChannel(channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get voice states"})
		return
	}

	c.JSON(http.StatusOK, states)
}
