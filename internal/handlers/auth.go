package handlers

import (
	"crypto/sha256"
	"net/http"
	"time"

	"discord-backend/config"
	"discord-backend/internal/database"
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

	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = generateDeviceID(c)
	}

	token, err := h.generateToken(user.ID, user.Email, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.saveToken(user.ID, deviceID, c.ClientIP(), token)

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

	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		deviceID = generateDeviceID(c)
	}

	deviceName := c.GetHeader("X-Device-Name")
	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	var user *models.User

	userService := services.NewUserService()
	dbUser, err := userService.GetByEmail(req.Email)
	if err == nil && dbUser != nil {
		if !userService.ValidatePassword(dbUser, req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		user = dbUser
	} else {
		builtInUser := services.GetBuiltInUserByEmail(req.Email)
		if builtInUser != nil {
			if !services.ValidateBuiltInPassword(builtInUser, req.Password) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			user = builtInUser
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
	}

	h.kickAllOtherDevices(user.ID)

	token, err := h.generateToken(user.ID, user.Email, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.saveToken(user.ID, deviceID, c.ClientIP(), token)

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	deviceID := middleware.GetDeviceID(c)

	if deviceID == "" {
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	_, err := database.DB.Exec(`
		DELETE FROM user_tokens WHERE user_id = $1 AND device_id = $2
	`, userID, deviceID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (h *AuthHandler) generateToken(userID uuid.UUID, email string, deviceID string) (string, error) {
	claims := &middleware.Claims{
		UserID:   userID,
		Email:    email,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(h.cfg.JWTExpireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func (h *AuthHandler) saveToken(userID uuid.UUID, deviceID string, ipAddress string, tokenString string) {
	tokenHash := hashToken(tokenString)
	expiresAt := time.Now().Add(time.Hour * time.Duration(h.cfg.JWTExpireHours))

	deviceName := "Unknown Device"

	database.DB.Exec(`
		INSERT INTO user_tokens (user_id, device_id, device_name, ip_address, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, device_id) DO UPDATE SET
			token_hash = EXCLUDED.token_hash,
			expires_at = EXCLUDED.expires_at,
			ip_address = EXCLUDED.ip_address
	`, userID, deviceID, deviceName, ipAddress, tokenHash, expiresAt)
}

func (h *AuthHandler) kickOtherDevices(userID uuid.UUID, currentDeviceID string) {
	_, err := database.DB.Exec(`
		DELETE FROM user_tokens WHERE user_id = $1 AND device_id != $2
	`, userID, currentDeviceID)
	if err != nil {
		return
	}
}

func (h *AuthHandler) kickAllOtherDevices(userID uuid.UUID) {
	_, err := database.DB.Exec(`
		DELETE FROM user_tokens WHERE user_id = $1
	`, userID)
	if err != nil {
		return
	}
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

func generateDeviceID(c *gin.Context) string {
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	raw := ip + userAgent
	hash := sha256.Sum256([]byte(raw))
	var result uuid.UUID
	copy(result[:], hash[:16])
	return result.String()
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	var result uuid.UUID
	copy(result[:], hash[:16])
	return result.String()
}
