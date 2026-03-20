package handlers

import (
	"net/http"

	"discord-backend/internal/middleware"
	"discord-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InviteHandler struct{}

func NewInviteHandler() *InviteHandler {
	return &InviteHandler{}
}

// POST /api/guilds/:id/invites — 生成邀请码
func (h *InviteHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	guildID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid guild ID"})
		return
	}

	var req struct {
		MaxUses     int `json:"max_uses"`
		ExpireHours int `json:"expire_hours"`
	}
	_ = c.ShouldBindJSON(&req)

	guildService := services.NewGuildService()
	if _, err := guildService.GetByID(guildID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
		return
	}

	inviteService := services.NewInviteService()
	code, err := inviteService.Generate(guildID, userID, req.MaxUses, req.ExpireHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invite"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": code})
}

// GET /api/invites/:code — 预览服务器（无需登录）
func (h *InviteHandler) Preview(c *gin.Context) {
	code := c.Param("code")

	inviteService := services.NewInviteService()
	guild, memberCount, err := inviteService.GetGuildByCode(code)
	if err != nil {
		h.respondInviteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"guild": gin.H{
			"id":           guild.ID,
			"name":         guild.Name,
			"icon":         guild.Icon,
			"member_count": memberCount,
		},
	})
}

// POST /api/invites/:code — 使用邀请码加入服务器
func (h *InviteHandler) Join(c *gin.Context) {
	userID := middleware.GetUserID(c)
	code := c.Param("code")

	inviteService := services.NewInviteService()
	guild, _, err := inviteService.Use(code, userID)
	if err != nil {
		h.respondInviteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"guild": gin.H{
			"id":   guild.ID,
			"name": guild.Name,
			"icon": guild.Icon,
		},
	})
}

func (h *InviteHandler) respondInviteError(c *gin.Context, err error) {
	switch err.Error() {
	case "invite code expired":
		c.JSON(http.StatusGone, gin.H{"error": "邀请链接已过期"})
	case "invite code has reached max uses":
		c.JSON(http.StatusGone, gin.H{"error": "邀请链接已达到最大使用次数"})
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "无效的邀请链接"})
	}
}
