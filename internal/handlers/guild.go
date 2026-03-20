package handlers

import (
	"net/http"

	"discord-backend/internal/middleware"
	"discord-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GuildHandler struct{}

func NewGuildHandler() *GuildHandler {
	return &GuildHandler{}
}

func (h *GuildHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Name string `json:"name" binding:"required,min=2,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	guildService := services.NewGuildService()
	guild, err := guildService.Create(req.Name, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guild"})
		return
	}

	c.JSON(http.StatusCreated, guild)
}

func (h *GuildHandler) Get(c *gin.Context) {
	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid guild ID"})
		return
	}

	guildService := services.NewGuildService()
	guild, err := guildService.GetByID(guildID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
		return
	}

	c.JSON(http.StatusOK, guild)
}

func (h *GuildHandler) GetUserGuilds(c *gin.Context) {
	userID := middleware.GetUserID(c)

	guildService := services.NewGuildService()
	guilds, err := guildService.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get guilds"})
		return
	}

	c.JSON(http.StatusOK, guilds)
}

func (h *GuildHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid guild ID"})
		return
	}

	guildService := services.NewGuildService()
	guild, err := guildService.GetByID(guildID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
		return
	}

	if guild.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only owner can delete guild"})
		return
	}

	if err := guildService.Delete(guildID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete guild"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Guild deleted"})
}

func (h *GuildHandler) Leave(c *gin.Context) {
	userID := middleware.GetUserID(c)
	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid guild ID"})
		return
	}

	guildService := services.NewGuildService()
	guild, err := guildService.GetByID(guildID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
		return
	}

	if guild.OwnerID == userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Owner cannot leave guild, delete it instead"})
		return
	}

	if err := guildService.RemoveMember(guildID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave guild"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Left guild successfully"})
}

func (h *GuildHandler) GetMembers(c *gin.Context) {
	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid guild ID"})
		return
	}

	guildService := services.NewGuildService()
	members, err := guildService.GetMembers(guildID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get members"})
		return
	}

	c.JSON(http.StatusOK, members)
}
