package handlers

import (
	"errors"
	"net/http"
	"time"

	"discord-backend/config"
	"discord-backend/internal/middleware"
	"discord-backend/internal/models"
	"discord-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=2,max=32"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userService := services.NewUserService()
	
	existingUser, _ := userService.GetByUsername(req.Username)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
		return
	}

	existingUser, _ = userService.GetByEmail(req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	user, err := userService.Create(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userService := services.NewUserService()
	user, err := userService.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !userService.ValidatePassword(user, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) generateToken(userID uuid.UUID, email string) (string, error) {
	claims := &middleware.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(h.cfg.JWTExpireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userService := services.NewUserService()

	user, err := userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

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
		Name     string              `json:"name" binding:"required,min=2,max=100"`
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
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channelService := services.NewChannelService()
	if err := channelService.Delete(channelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted"})
}

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

	c.JSON(http.StatusOK, messages)
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

func parseInt(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid character")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
