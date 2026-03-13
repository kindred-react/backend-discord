package router

import (
	"discord-backend/config"
	"discord-backend/internal/handlers"
	"discord-backend/internal/middleware"
	"discord-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, hub *websocket.Hub) *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", "::1", "localhost"})

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/ws", websocket.HandleWebSocket(hub))

	authHandler := handlers.NewAuthHandler(cfg)
	guildHandler := handlers.NewGuildHandler()
	channelHandler := handlers.NewChannelHandler()
	messageHandler := handlers.NewMessageHandler()
	voiceHandler := handlers.NewVoiceHandler()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.AuthMiddleware(cfg), authHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.GetMe)
		}

		guilds := api.Group("/guilds")
		guilds.Use(middleware.AuthMiddleware(cfg))
		{
			guilds.POST("", guildHandler.Create)
			guilds.GET("", guildHandler.GetUserGuilds)
			guilds.GET("/:id", guildHandler.Get)
			guilds.DELETE("/:id", guildHandler.Delete)
			guilds.GET("/:id/members", guildHandler.GetMembers)
			guilds.POST("/:id/channels", channelHandler.CreateByGuild)
			guilds.GET("/:id/channels", channelHandler.GetByGuild)
		}

		channels := api.Group("/channels")
		channels.Use(middleware.AuthMiddleware(cfg))
		{
			channels.GET("", channelHandler.GetAll)
			channels.GET("/:id", channelHandler.Get)
			channels.DELETE("/:id", channelHandler.Delete)
			channels.GET("/:id/messages", messageHandler.Get)
			channels.POST("/:id/messages", messageHandler.Create)
		}

		voice := api.Group("/voice")
		voice.Use(middleware.AuthMiddleware(cfg))
		{
			voice.POST("/join", voiceHandler.Join)
			voice.POST("/leave", voiceHandler.Leave)
			voice.GET("/:channelId", voiceHandler.GetChannelState)
		}
	}

	return r
}
