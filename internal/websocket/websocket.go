package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"discord-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	HandshakeTimeout: 10 * time.Second,
}

type Hub struct {
	clients    map[uuid.UUID]*Client
	rooms      map[string]map[*Client]bool
	broadcast  chan *BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

type Client struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Conn      *websocket.Conn
	Send      chan []byte
	hub       *Hub
}

type BroadcastMessage struct {
	Event     string
	ChannelID *uuid.UUID
	GuildID   *uuid.UUID
	Data      interface{}
}

type Message struct {
	Type      string      `json:"type"`
	Event     string      `json:"event"`
	ChannelID string      `json:"channel_id,omitempty"`
	GuildID   string      `json:"guild_id,omitempty"`
	Data      interface{} `json:"data"`
}

type SendMessagePayload struct {
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

type VoiceStatePayload struct {
	ChannelID string `json:"channel_id,omitempty"`
	GuildID   string `json:"guild_id,omitempty"`
}

type VoiceStateMessage struct {
	UserID    string  `json:"user_id"`
	ChannelID *string `json:"channel_id,omitempty"`
	GuildID   *string `json:"guild_id,omitempty"`
	Muted     bool    `json:"muted"`
	Deaf      bool    `json:"deaf"`
	SelfMute  bool    `json:"self_mute"`
	SelfDeaf  bool    `json:"self_deaf"`
}

var GlobalHub *Hub

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan *BroadcastMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.ID] = client
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.removeClient(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// Broadcast 公共方法，用于从其他地方发送广播消息
func (h *Hub) Broadcast(message *BroadcastMessage) {
	h.broadcast <- message
}

func (h *Hub) removeClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.Send)
		
		// Also remove from all rooms
		for roomName, clients := range h.rooms {
			if _, ok := clients[client]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.rooms, roomName)
				}
			}
		}
	}
}

func (h *Hub) handleBroadcast(message *BroadcastMessage) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var channelRoom, guildRoom string
	switch message.Event {
	case "message:new":
		if message.ChannelID != nil {
			channelRoom = "channel:" + message.ChannelID.String()
		}
	case "channel:voice:join", "channel:voice:leave":
		if message.GuildID != nil {
			guildRoom = "guild:" + message.GuildID.String()
		}
	}

	data, _ := json.Marshal(map[string]interface{}{
		"event":       message.Event,
		"data":        message.Data,
		"channel_id":  message.ChannelID,
		"guild_id":    message.GuildID,
	})

	if channelRoom != "" {
		log.Printf("Broadcast: event=%s, room=%s", message.Event, channelRoom)
		clients, ok := h.rooms[channelRoom]
		if ok {
			log.Printf("Broadcast: sending to %d clients in room %s", len(clients), channelRoom)
			for client := range clients {
				select {
				case client.Send <- data:
				default:
					delete(clients, client)
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			if len(clients) == 0 {
				delete(h.rooms, channelRoom)
			}
		} else {
			log.Printf("Broadcast: no clients in room %s", channelRoom)
		}
	}

	if guildRoom != "" {
		log.Printf("Broadcast: event=%s, room=%s", message.Event, guildRoom)
		clients, ok := h.rooms[guildRoom]
		if ok {
			log.Printf("Broadcast: sending to %d clients in room %s", len(clients), guildRoom)
			for client := range clients {
				select {
				case client.Send <- data:
				default:
					delete(clients, client)
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			if len(clients) == 0 {
				delete(h.rooms, guildRoom)
			}
		} else {
			log.Printf("Broadcast: no clients in room %s", guildRoom)
		}
	}
}

func (h *Hub) JoinRoom(client *Client, room string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
}

func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.rooms[room] != nil {
		delete(h.rooms[room], client)
	}
}

func (h *Hub) GetRoomClients(room string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.rooms[room] != nil {
		return len(h.rooms[room])
	}
	return 0
}

func (h *Hub) GetChannelClients(channelID uuid.UUID) []*Client {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	room := "channel:" + channelID.String()
	var clients []*Client
	if roomClients, ok := h.rooms[room]; ok {
		for client := range roomClients {
			clients = append(clients, client)
		}
	}
	return clients
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("ReadPump: connection error: %v", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *Client) handleMessage(msg Message) {
	log.Printf("Received message: event=%s, data=%v", msg.Event, msg.Data)
	switch msg.Event {
	case "message:send":
		var payload SendMessagePayload
		data, _ := json.Marshal(msg.Data)
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Printf("Failed to unmarshal message payload: %v", err)
			return
		}

		log.Printf("Processing message: channel=%s, content=%s", payload.ChannelID, payload.Content)

		channelID, _ := uuid.Parse(payload.ChannelID)
		messageService := services.NewMessageService()
		message, err := messageService.Create(channelID, c.UserID, payload.Content, nil)
		if err != nil {
			log.Printf("Failed to create message: %v", err)
			return
		}

		channelService := services.NewChannelService()
		channel, err := channelService.GetByID(channelID)
		if err != nil {
			log.Printf("Failed to get channel: %v", err)
			return
		}

		// 广播消息到所有在该频道的客户端
		c.hub.broadcast <- &BroadcastMessage{
			Event:     "message:new",
			ChannelID: &channelID,
			GuildID:   channel.GuildID,
			Data:      message,
		}

	case "channel:voice:join":
		var payload VoiceStatePayload
		data, _ := json.Marshal(msg.Data)
		if err := json.Unmarshal(data, &payload); err != nil {
			return
		}

		channelID, _ := uuid.Parse(payload.ChannelID)
		guildID, _ := uuid.Parse(payload.GuildID)

		voiceService := services.NewVoiceService()
		state, err := voiceService.JoinChannel(c.UserID, channelID, guildID)
		if err != nil {
			log.Printf("Failed to join voice channel: %v", err)
			return
		}

		channelIDStr := state.ChannelID.String()
		guildIDStr := state.GuildID.String()
		voiceMsg := VoiceStateMessage{
			UserID:    c.UserID.String(),
			ChannelID: &channelIDStr,
			GuildID:   &guildIDStr,
		}

		c.hub.broadcast <- &BroadcastMessage{
			Event:   "channel:voice:join",
			GuildID: state.GuildID,
			Data:    voiceMsg,
		}

		room := "guild:" + guildIDStr
		c.hub.JoinRoom(c, room)

	case "channel:voice:leave":
		var payload VoiceStatePayload
		data, _ := json.Marshal(msg.Data)
		if err := json.Unmarshal(data, &payload); err != nil {
			return
		}

		guildID, _ := uuid.Parse(payload.GuildID)

		voiceService := services.NewVoiceService()
		if err := voiceService.LeaveChannel(c.UserID); err != nil {
			log.Printf("Failed to leave voice channel: %v", err)
			return
		}

		guildIDStr := guildID.String()
		voiceMsg := VoiceStateMessage{
			UserID:  c.UserID.String(),
			GuildID: &guildIDStr,
		}

		room := "guild:" + guildIDStr
		c.hub.LeaveRoom(c, room)

		c.hub.broadcast <- &BroadcastMessage{
			Event:   "channel:voice:leave",
			GuildID: &guildID,
			Data:    voiceMsg,
		}

	case "join_channel":
		var payload struct {
			ChannelID string `json:"channel_id"`
		}
		data, _ := json.Marshal(msg.Data)
		if err := json.Unmarshal(data, &payload); err != nil {
			return
		}

		log.Printf("User %s joining channel room: %s", c.UserID, payload.ChannelID)

		room := "channel:" + payload.ChannelID
		c.hub.JoinRoom(c, room)

		log.Printf("User %s joined room %s, room now has clients: %d", c.UserID, room, c.hub.GetRoomClients(room))

		channelID, err := uuid.Parse(payload.ChannelID)
		if err != nil {
			log.Printf("Failed to parse channel ID: %v", err)
			return
		}

		messageService := services.NewMessageService()
		messages, err := messageService.GetByChannelID(channelID, 50, 0)
		if err != nil {
			log.Printf("Failed to get channel messages: %v", err)
			return
		}

		historyData, _ := json.Marshal(map[string]interface{}{
			"event":   "message:history",
			"data":    messages,
			"channel_id": payload.ChannelID,
		})

		select {
		case c.Send <- historyData:
			log.Printf("Sent %d historical messages to user %s for channel %s", len(messages), c.UserID, payload.ChannelID)
		default:
			log.Printf("Failed to send historical messages to user %s", c.UserID)
		}

	case "leave_channel":
		var payload struct {
			ChannelID string `json:"channel_id"`
		}
		data, _ := json.Marshal(msg.Data)
		if err := json.Unmarshal(data, &payload); err != nil {
			return
		}

		room := "channel:" + payload.ChannelID
		c.hub.LeaveRoom(c, room)
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("WritePump: error getting writer: %v", err)
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				log.Printf("WritePump: error closing writer: %v", err)
				return
			}

		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WritePump: ping error: %v", err)
				return
			}
		}
	}
}

func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("WebSocket handler: starting")
		userIDStr := c.Query("user_id")
		if userIDStr == "" {
			log.Println("WebSocket handler: user_id is empty")
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Printf("WebSocket handler: invalid user_id: %s, error: %v", userIDStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}

		log.Printf("WebSocket handler: upgrading connection for user: %s", userID)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		log.Printf("WebSocket: connection upgraded successfully for user: %s", userID)

		client := &Client{
			ID:     uuid.New(),
			UserID: userID,
			Conn:   conn,
			Send:   make(chan []byte, 256),
			hub:    hub,
		}

		hub.register <- client

		channelService := services.NewChannelService()
		guildService := services.NewGuildService()
		messageService := services.NewMessageService()
		userGuilds, err := guildService.GetByUserID(userID)
		if err != nil {
			log.Printf("Failed to get user guilds: %v", err)
		} else {
			for _, guild := range userGuilds {
				channels, err := channelService.GetByGuildID(guild.ID)
				if err != nil {
					log.Printf("Failed to get channels for guild %s: %v", guild.ID, err)
					continue
				}
				for _, channel := range channels {
					room := "channel:" + channel.ID.String()
					hub.JoinRoom(client, room)
					log.Printf("User %s auto-joined channel room: %s (channel: %s)", userID, room, channel.Name)

					if channel.Type == "text" {
						messages, err := messageService.GetByChannelID(channel.ID, 50, 0)
						if err != nil {
							log.Printf("Failed to get historical messages for channel %s: %v", channel.ID, err)
							continue
						}

						historyData, _ := json.Marshal(map[string]interface{}{
							"event":      "message:history",
							"data":       messages,
							"channel_id": channel.ID.String(),
						})

						select {
						case client.Send <- historyData:
							log.Printf("Sent %d historical messages to user %s for channel %s", len(messages), userID, channel.Name)
						default:
							log.Printf("Failed to send historical messages to user %s for channel %s", userID, channel.Name)
						}
					}
				}
			}
		}

		go client.WritePump()
		client.ReadPump()
	}
}
