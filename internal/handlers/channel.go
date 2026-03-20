package handlers

import (
	"net/http"

	"discord-backend/internal/middleware"
	"discord-backend/internal/models"
	"discord-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChannelHandler struct{}

func NewChannelHandler() *ChannelHandler {
	return &ChannelHandler{}
}

func (h *ChannelHandler) GetAll(c *gin.Context) {
	channelService := services.NewChannelService()
	channels, err := channelService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channels"})
		return
	}

	c.JSON(http.StatusOK, channels)
}

func (h *ChannelHandler) CreateByGuild(c *gin.Context) {
	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid guild ID"})
		return
	}

	var req struct {
		Name     string              `json:"name" binding:"required,min=1,max=100"`
		Type     models.ChannelType `json:"type" binding:"required"`
		ParentID *uuid.UUID          `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channelService := services.NewChannelService()
	channel, err := channelService.Create(req.Name, req.Type, guildID, req.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

func (h *ChannelHandler) Get(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channelService := services.NewChannelService()
	channel, err := channelService.GetByID(channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

func (h *ChannelHandler) GetByGuild(c *gin.Context) {
	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid guild ID"})
		return
	}

	channelService := services.NewChannelService()
	channels, err := channelService.GetByGuildID(guildID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get channels"})
		return
	}

	c.JSON(http.StatusOK, channels)
}

func (h *ChannelHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channelService := services.NewChannelService()
	channel, err := channelService.GetByID(channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	// Check if user is owner or admin of the guild
	guildService := services.NewGuildService()
	guildID := channel.GuildID
	if guildID == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Channel has no guild"})
		return
	}
	guild, err := guildService.GetByID(*guildID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
		return
	}

	if guild.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only guild owner can delete channels"})
		return
	}

	if err := channelService.Delete(channelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted"})
}
