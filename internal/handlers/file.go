package handlers

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"discord-backend/internal/middleware"
	"discord-backend/internal/models"
	"discord-backend/internal/services"
	"discord-backend/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	hub *websocket.Hub
}

func NewFileHandler(hub *websocket.Hub) *FileHandler {
	return &FileHandler{
		hub: hub,
	}
}

// UploadFile 处理文件上传
func (h *FileHandler) UploadFile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	channelID, err := uuid.Parse(c.PostForm("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 检查文件大小 (限制为 10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 10MB)"})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	
	// 确保上传目录存在
	uploadDir := "./uploads/files"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// 保存文件
	filepath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 创建消息记录 - 使用 file 类型
	fileURL := fmt.Sprintf("/uploads/files/%s", filename)
	content := file.Filename
	
	// 创建附件信息
	attachments := []models.Attachment{
		{
			ID:          uuid.New().String(),
			Filename:    file.Filename,
			URL:         fileURL,
			ProxyURL:    fileURL,
			Size:        file.Size,
			ContentType: file.Header.Get("Content-Type"),
		},
	}
	
	messageService := services.NewMessageService()
	message, err := messageService.CreateWithAttachments(channelID, userID, content, "file", &fileURL, nil, attachments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
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
		"message":  message,
		"file_url": fileURL,
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// UploadImage 处理图片上传
func (h *FileHandler) UploadImage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	channelID, err := uuid.Parse(c.PostForm("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image format"})
		return
	}

	// 检查文件大小 (限制为 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image too large (max 5MB)"})
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	
	// 确保上传目录存在
	uploadDir := "./uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// 保存文件
	filePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	// 获取图片尺寸
	var width, height int
	if imgFile, err := os.Open(filePath); err == nil {
		defer imgFile.Close()
		if img, _, err := image.DecodeConfig(imgFile); err == nil {
			width = img.Width
			height = img.Height
		}
	}

	// 创建消息记录 - 使用 image 类型
	imageURL := fmt.Sprintf("/uploads/images/%s", filename)
	content := file.Filename
	
	// 创建附件信息（包含图片尺寸）
	attachments := []models.Attachment{
		{
			ID:          uuid.New().String(),
			Filename:    file.Filename,
			URL:         imageURL,
			ProxyURL:    imageURL,
			Size:        file.Size,
			Width:       &width,
			Height:      &height,
			ContentType: file.Header.Get("Content-Type"),
		},
	}
	
	messageService := services.NewMessageService()
	message, err := messageService.CreateWithAttachments(channelID, userID, content, "image", &imageURL, nil, attachments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
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
		"message":   message,
		"image_url": imageURL,
		"filename":  file.Filename,
		"size":      file.Size,
	})
}

// SendGif 发送 GIF
func (h *FileHandler) SendGif(c *gin.Context) {
	userID := middleware.GetUserID(c)
	
	var req struct {
		ChannelID string `json:"channel_id" binding:"required"`
		GifURL    string `json:"gif_url" binding:"required"`
		Title     string `json:"title"`
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

	// 创建 GIF 消息
	content := req.GifURL
	if req.Title != "" {
		content = req.Title
	}

	messageService := services.NewMessageService()
	message, err := messageService.CreateWithType(channelID, userID, content, "gif", &req.GifURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send GIF"})
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

	c.JSON(http.StatusOK, message)
}

// SendSticker 发送贴纸
func (h *FileHandler) SendSticker(c *gin.Context) {
	userID := middleware.GetUserID(c)
	
	var req struct {
		ChannelID string `json:"channel_id" binding:"required"`
		Sticker   string `json:"sticker" binding:"required"`
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

	// 创建贴纸消息
	messageService := services.NewMessageService()
	message, err := messageService.CreateWithType(channelID, userID, req.Sticker, "sticker", nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send sticker"})
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

	c.JSON(http.StatusOK, message)
}

// DownloadFile 下载文件
func (h *FileHandler) DownloadFile(c *gin.Context) {
	filename := c.Param("filename")
	fileType := c.Param("type") // "files" or "images"

	// 验证文件类型
	if fileType != "files" && fileType != "images" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}

	// 构建文件路径
	filepath := filepath.Join("./uploads", fileType, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 打开文件
	file, err := os.Open(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info"})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// 发送文件
	io.Copy(c.Writer, file)
}
