package middleware

import (
	"net/http"
	"strings"

	"discord-backend/config"
	"discord-backend/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	DeviceID string    `json:"device_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims.DeviceID != "" {
			var exists bool
			err := database.DB.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM user_tokens 
					WHERE user_id = $1 AND device_id = $2 AND expires_at > CURRENT_TIMESTAMP
				)
			`, claims.UserID, claims.DeviceID).Scan(&exists)

			if err != nil || !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired or logged out from another device"})
				c.Abort()
				return
			}
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("deviceID", claims.DeviceID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("userID")
	if !exists {
		return uuid.Nil
	}
	return userID.(uuid.UUID)
}

func GetDeviceID(c *gin.Context) string {
	deviceID, exists := c.Get("deviceID")
	if !exists {
		return ""
	}
	return deviceID.(string)
}
