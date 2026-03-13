package services

import (
	"database/sql"
	"encoding/json"

	"discord-backend/internal/database"
	"discord-backend/internal/models"

	"github.com/google/uuid"
)

type GuildService struct{}

func NewGuildService() *GuildService {
	return &GuildService{}
}

func (s *GuildService) Create(name string, ownerID uuid.UUID) (*models.Guild, error) {
	var guild models.Guild
	err := database.DB.QueryRow(`
		INSERT INTO guilds (name, owner_id)
		VALUES ($1, $2)
		RETURNING id, name, icon, owner_id, created_at, updated_at
	`, name, ownerID).Scan(&guild.ID, &guild.Name, &guild.Icon, &guild.OwnerID,
		&guild.CreatedAt, &guild.UpdatedAt)

	if err != nil {
		return nil, err
	}

	_, err = database.DB.Exec(`
		INSERT INTO guild_members (guild_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, guild.ID, ownerID)

	if err != nil {
		return nil, err
	}

	defaultChannels := []struct {
		name  string
		ctype string
	}{
		{"general", "text"},
		{"General", "text"},
		{"Voice", "voice"},
	}

	for i, ch := range defaultChannels {
		_, err = database.DB.Exec(`
			INSERT INTO channels (name, type, guild_id, position)
			VALUES ($1, $2, $3, $4)
		`, ch.name, ch.ctype, guild.ID, i)
		if err != nil {
			return nil, err
		}
	}

	return &guild, nil
}

func (s *GuildService) GetByID(id uuid.UUID) (*models.Guild, error) {
	var guild models.Guild
	err := database.DB.QueryRow(`
		SELECT id, name, icon, owner_id, created_at, updated_at
		FROM guilds WHERE id = $1
	`, id).Scan(&guild.ID, &guild.Name, &guild.Icon, &guild.OwnerID,
		&guild.CreatedAt, &guild.UpdatedAt)

	if err != nil {
		return nil, err
	}

	owner, _ := NewUserService().GetByID(guild.OwnerID)
	guild.Owner = owner

	channels, _ := NewChannelService().GetByGuildID(id)
	guild.Channels = channels

	members, _ := s.GetMembers(id)
	guild.Members = members

	return &guild, nil
}

func (s *GuildService) GetByUserID(userID uuid.UUID) ([]*models.Guild, error) {
	rows, err := database.DB.Query(`
		SELECT g.id, g.name, g.icon, g.owner_id, g.created_at, g.updated_at
		FROM guilds g
		JOIN guild_members gm ON g.id = gm.guild_id
		WHERE gm.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guilds []*models.Guild
	for rows.Next() {
		var guild models.Guild
		err := rows.Scan(&guild.ID, &guild.Name, &guild.Icon, &guild.OwnerID,
			&guild.CreatedAt, &guild.UpdatedAt)
		if err != nil {
			return nil, err
		}
		guilds = append(guilds, &guild)
	}

	return guilds, nil
}

func (s *GuildService) GetMembers(guildID uuid.UUID) ([]*models.GuildMember, error) {
	rows, err := database.DB.Query(`
		SELECT gm.guild_id, gm.user_id, gm.nickname, gm.role, gm.joined_at,
			   u.id, u.username, u.email, u.avatar, u.discriminator
		FROM guild_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.guild_id = $1
	`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.GuildMember
	for rows.Next() {
		var member models.GuildMember
		var user models.User
		err := rows.Scan(&member.GuildID, &member.UserID, &member.Nickname,
			&member.Role, &member.JoinedAt,
			&user.ID, &user.Username, &user.Email, &user.Avatar, &user.Discriminator)
		if err != nil {
			return nil, err
		}
		member.User = &user
		members = append(members, &member)
	}

	return members, nil
}

func (s *GuildService) AddMember(guildID, userID uuid.UUID) error {
	_, err := database.DB.Exec(`
		INSERT INTO guild_members (guild_id, user_id, role)
		VALUES ($1, $2, 'member')
		ON CONFLICT DO NOTHING
	`, guildID, userID)
	return err
}

func (s *GuildService) RemoveMember(guildID, userID uuid.UUID) error {
	_, err := database.DB.Exec(`
		DELETE FROM guild_members WHERE guild_id = $1 AND user_id = $2
	`, guildID, userID)
	return err
}

func (s *GuildService) Delete(id uuid.UUID) error {
	_, err := database.DB.Exec(`DELETE FROM guilds WHERE id = $1`, id)
	return err
}

type ChannelService struct{}

func NewChannelService() *ChannelService {
	return &ChannelService{}
}

func (s *ChannelService) Create(name string, channelType models.ChannelType, guildID uuid.UUID, parentID *uuid.UUID) (*models.Channel, error) {
	var channel models.Channel
	err := database.DB.QueryRow(`
		INSERT INTO channels (name, type, guild_id, parent_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, type, guild_id, parent_id, position, topic, created_at, updated_at
	`, name, channelType, guildID, parentID).Scan(
		&channel.ID, &channel.Name, &channel.Type, &channel.GuildID,
		&channel.ParentID, &channel.Position, &channel.Topic,
		&channel.CreatedAt, &channel.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (s *ChannelService) GetByID(id uuid.UUID) (*models.Channel, error) {
	var channel models.Channel
	err := database.DB.QueryRow(`
		SELECT id, name, type, guild_id, parent_id, position, topic, created_at, updated_at
		FROM channels WHERE id = $1
	`, id).Scan(&channel.ID, &channel.Name, &channel.Type, &channel.GuildID,
		&channel.ParentID, &channel.Position, &channel.Topic,
		&channel.CreatedAt, &channel.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (s *ChannelService) GetAll() ([]*models.Channel, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, type, guild_id, parent_id, position, topic, created_at, updated_at
		FROM channels ORDER BY position
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*models.Channel
	for rows.Next() {
		var channel models.Channel
		err := rows.Scan(
			&channel.ID, &channel.Name, &channel.Type, &channel.GuildID,
			&channel.ParentID, &channel.Position, &channel.Topic,
			&channel.CreatedAt, &channel.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, &channel)
	}

	return channels, nil
}

func (s *ChannelService) GetByGuildID(guildID uuid.UUID) ([]*models.Channel, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, type, guild_id, parent_id, position, topic, created_at, updated_at
		FROM channels WHERE guild_id = $1 ORDER BY position
	`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*models.Channel
	for rows.Next() {
		var channel models.Channel
		err := rows.Scan(&channel.ID, &channel.Name, &channel.Type, &channel.GuildID,
			&channel.ParentID, &channel.Position, &channel.Topic,
			&channel.CreatedAt, &channel.UpdatedAt)
		if err != nil {
			return nil, err
		}
		channels = append(channels, &channel)
	}

	return channels, nil
}

func (s *ChannelService) Update(id uuid.UUID, name string) error {
	_, err := database.DB.Exec(`
		UPDATE channels SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
	`, name, id)
	return err
}

func (s *ChannelService) Delete(id uuid.UUID) error {
	_, err := database.DB.Exec(`DELETE FROM channels WHERE id = $1`, id)
	return err
}

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (s *MessageService) Create(channelID, authorID uuid.UUID, content string, replyToID *uuid.UUID) (*models.Message, error) {
	var message models.Message
	var embedsJSON []byte
	err := database.DB.QueryRow(`
		INSERT INTO messages (channel_id, author_id, content, reply_to_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, channel_id, author_id, content, embeds, reply_to_id, created_at, updated_at
	`, channelID, authorID, content, replyToID).Scan(
		&message.ID, &message.ChannelID, &message.AuthorID, &message.Content,
		&embedsJSON, &message.ReplyToID, &message.CreatedAt, &message.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(embedsJSON, &message.Embeds); err != nil {
		message.Embeds = []models.Embed{}
	}

	author, _ := NewUserService().GetByID(authorID)
	message.Author = author

	channel, _ := NewChannelService().GetByID(channelID)
	message.Channel = channel

	return &message, nil
}

func (s *MessageService) GetByID(id uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := database.DB.QueryRow(`
		SELECT id, channel_id, author_id, content, embeds, reply_to_id, created_at, updated_at
		FROM messages WHERE id = $1
	`, id).Scan(&message.ID, &message.ChannelID, &message.AuthorID, &message.Content,
		&message.Embeds, &message.ReplyToID, &message.CreatedAt, &message.UpdatedAt)

	if err != nil {
		return nil, err
	}

	author, _ := NewUserService().GetByID(message.AuthorID)
	message.Author = author

	channel, _ := NewChannelService().GetByID(message.ChannelID)
	message.Channel = channel

	return &message, nil
}

func (s *MessageService) GetByChannelID(channelID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	rows, err := database.DB.Query(`
		SELECT id, channel_id, author_id, content, embeds, reply_to_id, created_at, updated_at
		FROM messages WHERE channel_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var message models.Message
		err := rows.Scan(&message.ID, &message.ChannelID, &message.AuthorID,
			&message.Content, &message.Embeds, &message.ReplyToID,
			&message.CreatedAt, &message.UpdatedAt)
		if err != nil {
			return nil, err
		}
		author, _ := NewUserService().GetByID(message.AuthorID)
		message.Author = author
		messages = append(messages, &message)
	}

	return messages, nil
}

func (s *MessageService) Delete(id uuid.UUID) error {
	_, err := database.DB.Exec(`DELETE FROM messages WHERE id = $1`, id)
	return err
}

type VoiceService struct{}

func NewVoiceService() *VoiceService {
	return &VoiceService{}
}

func (s *VoiceService) JoinChannel(userID, channelID, guildID uuid.UUID) (*models.VoiceState, error) {
	state := &models.VoiceState{
		UserID:    userID,
		ChannelID: &channelID,
		GuildID:   &guildID,
	}

	err := database.DB.QueryRow(`
		INSERT INTO voice_states (user_id, channel_id, guild_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			channel_id = $2,
			guild_id = $3,
			joined_at = CURRENT_TIMESTAMP
		RETURNING joined_at
	`, userID, channelID, guildID).Scan(&state.JoinedAt)

	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *VoiceService) LeaveChannel(userID uuid.UUID) error {
	_, err := database.DB.Exec(`
		UPDATE voice_states SET channel_id = NULL, guild_id = NULL WHERE user_id = $1
	`, userID)
	return err
}

func (s *VoiceService) GetByUser(userID uuid.UUID) (*models.VoiceState, error) {
	var state models.VoiceState
	err := database.DB.QueryRow(`
		SELECT user_id, channel_id, guild_id, joined_at, deaf, muted, self_deaf, self_muted
		FROM voice_states WHERE user_id = $1
	`, userID).Scan(&state.UserID, &state.ChannelID, &state.GuildID,
		&state.JoinedAt, &state.Deaf, &state.Muted, &state.SelfDeaf, &state.SelfMute)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (s *VoiceService) GetByChannel(channelID uuid.UUID) ([]*models.VoiceState, error) {
	rows, err := database.DB.Query(`
		SELECT user_id, channel_id, guild_id, joined_at, deaf, muted, self_deaf, self_muted
		FROM voice_states WHERE channel_id = $1
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []*models.VoiceState
	for rows.Next() {
		var state models.VoiceState
		err := rows.Scan(&state.UserID, &state.ChannelID, &state.GuildID,
			&state.JoinedAt, &state.Deaf, &state.Muted, &state.SelfDeaf, &state.SelfMute)
		if err != nil {
			return nil, err
		}
		states = append(states, &state)
	}

	return states, nil
}

func (s *VoiceService) UpdateMute(userID uuid.UUID, muted bool) error {
	_, err := database.DB.Exec(`
		UPDATE voice_states SET self_muted = $1 WHERE user_id = $2
	`, muted, userID)
	return err
}

func (s *VoiceService) UpdateDeaf(userID uuid.UUID, deaf bool) error {
	_, err := database.DB.Exec(`
		UPDATE voice_states SET self_deaf = $1 WHERE user_id = $2
	`, deaf, userID)
	return err
}
