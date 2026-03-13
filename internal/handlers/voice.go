package handlers

import (
	"fmt"
	"net/http"
	"time"

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

func (h *VoiceHandler) EndCall(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		ChannelID    uuid.UUID   `json:"channel_id" binding:"required"`
		Participants []uuid.UUID `json:"participants" binding:"required"`
		HasVideo     bool        `json:"has_video"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	voiceService := services.NewVoiceService()
	voiceCall, err := voiceService.EndCall(req.ChannelID, userID, req.Participants, req.HasVideo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end voice call"})
		return
	}

	c.JSON(http.StatusOK, voiceCall)
}

func (h *VoiceHandler) UploadVoiceMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	channelID, err := uuid.Parse(c.PostForm("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	file, err := c.FormFile("voice")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No voice file provided"})
		return
	}

	// 保存文件到本地或云存储
	filename := fmt.Sprintf("voice_%s_%d.webm", userID.String(), time.Now().Unix())
	filepath := fmt.Sprintf("./uploads/voice/%s", filename)

	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save voice file"})
		return
	}

	voiceURL := fmt.Sprintf("/uploads/voice/%s", filename)
	duration := c.PostForm("duration")

	messageService := services.NewMessageService()
	msg, err := messageService.CreateVoiceMessage(channelID, userID, voiceURL, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create voice message"})
		return
	}

	c.JSON(http.StatusOK, msg)
}
