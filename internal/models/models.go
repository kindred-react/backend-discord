package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Username      string     `json:"username" db:"username"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	Avatar        *string    `json:"avatar" db:"avatar"`
	Discriminator string     `json:"discriminator" db:"discriminator"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type Guild struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Icon        *string         `json:"icon" db:"icon"`
	OwnerID     uuid.UUID       `json:"owner_id" db:"owner_id"`
	Owner       *User           `json:"owner,omitempty"`
	Members     []*GuildMember  `json:"members,omitempty"`
	Channels    []*Channel      `json:"channels,omitempty"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type GuildMember struct {
	GuildID   uuid.UUID `json:"guild_id" db:"guild_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Nickname  *string  `json:"nickname" db:"nickname"`
	Role      string   `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
	User      *User    `json:"user,omitempty"`
	Guild     *Guild   `json:"guild,omitempty"`
}

type ChannelType string

const (
	ChannelTypeText   ChannelType = "text"
	ChannelTypeVoice  ChannelType = "voice"
	ChannelTypeCategory ChannelType = "category"
)

type Channel struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Type     ChannelType `json:"type" db:"type"`
	GuildID   *uuid.UUID `json:"guild_id" db:"guild_id"`
	Guild     *Guild     `json:"guild,omitempty"`
	ParentID  *uuid.UUID `json:"parent_id" db:"parent_id"`
	Position  int        `json:"position" db:"position"`
	Topic     *string    `json:"topic" db:"topic"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	ChannelID uuid.UUID  `json:"channel_id" db:"channel_id"`
	Channel   *Channel   `json:"channel,omitempty"`
	AuthorID  uuid.UUID  `json:"author_id" db:"author_id"`
	Author    *User      `json:"author,omitempty"`
	Content   string     `json:"content" db:"content"`
	Embeds    []Embed    `json:"embeds" db:"embeds"`
	ReplyToID *uuid.UUID `json:"reply_to_id" db:"reply_to_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

type Embed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Color       int    `json:"color"`
	Image       string `json:"image"`
	Thumbnail   string `json:"thumbnail"`
	Author      *EmbedAuthor `json:"author"`
	Fields      []EmbedField `json:"fields"`
	Footer      *EmbedFooter `json:"footer"`
	Timestamp   string `json:"timestamp"`
}

type EmbedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type EmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type VoiceState struct {
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	ChannelID *uuid.UUID `json:"channel_id" db:"channel_id"`
	GuildID   *uuid.UUID `json:"guild_id" db:"guild_id"`
	JoinedAt  time.Time  `json:"joined_at" db:"joined_at"`
	 Deaf      bool       `json:"deaf" db:"deaf"`
	Muted     bool       `json:"muted" db:"muted"`
	SelfDeaf  bool       `json:"self_deaf" db:"self_deaf"`
	SelfMute  bool       `json:"self_mute" db:"self_mute"`
}
