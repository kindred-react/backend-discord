package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"discord-backend/config"
)

var DB *sql.DB

func Initialize(cfg *config.Config) error {
	var err error
	DB, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

func CreateTables() error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		username VARCHAR(32) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		avatar VARCHAR(255),
		discriminator VARCHAR(4) NOT NULL DEFAULT '0001',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS guilds (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		icon VARCHAR(255),
		owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS guild_members (
		guild_id UUID NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		nickname VARCHAR(32),
		role VARCHAR(32) NOT NULL DEFAULT 'member',
		joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (guild_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS channels (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(100) NOT NULL,
		type VARCHAR(20) NOT NULL DEFAULT 'text',
		guild_id UUID REFERENCES guilds(id) ON DELETE CASCADE,
		parent_id UUID REFERENCES channels(id) ON DELETE SET NULL,
		position INTEGER NOT NULL DEFAULT 0,
		topic VARCHAR(255),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
		author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		content TEXT NOT NULL,
		embeds JSONB DEFAULT '[]',
		reply_to_id UUID REFERENCES messages(id) ON DELETE SET NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS voice_states (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		channel_id UUID REFERENCES channels(id) ON DELETE SET NULL,
		guild_id UUID REFERENCES guilds(id) ON DELETE CASCADE,
		joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deaf BOOLEAN DEFAULT FALSE,
		muted BOOLEAN DEFAULT FALSE,
		self_deaf BOOLEAN DEFAULT FALSE,
		self_muted BOOLEAN DEFAULT FALSE,
		PRIMARY KEY (user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_channel_id ON messages(channel_id);
	CREATE INDEX IF NOT EXISTS idx_messages_author_id ON messages(author_id);
	CREATE INDEX IF NOT EXISTS idx_guild_members_user_id ON guild_members(user_id);
	CREATE INDEX IF NOT EXISTS idx_channels_guild_id ON channels(guild_id);
	`
	
	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database tables created successfully")

	log.Println("Running seedDefaultData...")
	if err := seedDefaultData(); err != nil {
		log.Printf("Warning: failed to seed default data: %v", err)
	} else {
		log.Println("Seed data completed")
	}

	return nil
}

func seedDefaultData() error {
	log.Println("Checking channels table...")

	requiredChannels := []string{
		"6ba7b810-9dad-11d1-80b4-00c04fd430c1",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c2",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c3",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c4",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c5",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c6",
	}

	for _, chID := range requiredChannels {
		var exists bool
		err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM channels WHERE id = $1)", chID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check channel %s: %w", chID, err)
		}

		if !exists {
			log.Printf("Channel %s does not exist, need to insert", chID)
		}
	}

	channels := []struct {
		id   string
		name string
		typ  string
	}{
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c1", "欢迎", "text"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c2", "规则", "text"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c3", "公告", "text"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c4", "综合", "text"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c5", "语音聊天", "voice"},
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c6", "音乐", "voice"},
	}

	for i, ch := range channels {
		result, err := DB.Exec(`
			INSERT INTO channels (id, name, type, position)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, type = EXCLUDED.type
		`, ch.id, ch.name, ch.typ, i)
		if err != nil {
			return fmt.Errorf("failed to insert channel %s: %w", ch.name, err)
		}
		rowsAffected, _ := result.RowsAffected()
		log.Printf("Inserted/updated channel %s, rows affected: %d", ch.name, rowsAffected)
	}

	log.Println("Seeding default users...")
	users := []struct {
		id       string
		username string
		email    string
		password string
	}{
		{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "甲", "jia@test.com", "password123"},
		{"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "乙", "yi@test.com", "password123"},
		{"cccccccc-cccc-cccc-cccc-cccccccccccc", "丙", "bing@test.com", "password123"},
	}

	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Warning: failed to hash password for %s: %v", u.username, err)
			continue
		}
		_, err = DB.Exec(`
			INSERT INTO users (id, username, email, password_hash, discriminator)
			VALUES ($1, $2, $3, $4, '0001')
			ON CONFLICT (username) DO UPDATE SET email = EXCLUDED.email
		`, u.id, u.username, u.email, string(hash))
		if err != nil {
			log.Printf("Warning: failed to insert user %s: %v", u.username, err)
		} else {
			log.Printf("Inserted/updated user: %s", u.username)
		}
	}

	log.Println("Seeding default guilds...")
	guilds := []struct {
		id      string
		name    string
		ownerID string
	}{
		{"11111111-1111-1111-1111-111111111111", "A", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"},
		{"22222222-2222-2222-2222-222222222222", "B", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"},
		{"33333333-3333-3333-3333-333333333333", "C", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"},
	}

	for _, g := range guilds {
		_, err := DB.Exec(`
			INSERT INTO guilds (id, name, owner_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
		`, g.id, g.name, g.ownerID)
		if err != nil {
			log.Printf("Warning: failed to insert guild %s: %v", g.name, err)
		} else {
			log.Printf("Inserted/updated guild: %s", g.name)
		}
	}

	log.Println("Seeding guild members...")
	allUsers := []string{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "cccccccc-cccc-cccc-cccc-cccccccccccc"}
	allGuilds := []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "33333333-3333-3333-3333-333333333333"}

	for _, guildID := range allGuilds {
		for _, userID := range allUsers {
			_, err := DB.Exec(`
				INSERT INTO guild_members (guild_id, user_id, role)
				VALUES ($1, $2, 'member')
				ON CONFLICT (guild_id, user_id) DO NOTHING
			`, guildID, userID)
			if err != nil {
				log.Printf("Warning: failed to add member to guild: %v", err)
			}
		}
	}

	log.Println("Seeding guild channels...")
	guildChannels := []struct {
		id      string
		name    string
		typ     string
		guildID string
		position int
	}{
		{"aaaaaaaa-aaaa-aaaa-aaaa-111111111111", "欢迎", "text", "11111111-1111-1111-1111-111111111111", 0},
		{"aaaaaaaa-aaaa-aaaa-aaaa-111111111112", "综合", "text", "11111111-1111-1111-1111-111111111111", 1},
		{"aaaaaaaa-aaaa-aaaa-aaaa-111111111113", "语音1", "voice", "11111111-1111-1111-1111-111111111111", 2},
		{"aaaaaaaa-aaaa-aaaa-aaaa-111111111114", "语音2", "voice", "11111111-1111-1111-1111-111111111111", 3},
		
		{"bbbbbbbb-bbbb-bbbb-bbbb-222222222221", "公告", "text", "22222222-2222-2222-2222-222222222222", 0},
		{"bbbbbbbb-bbbb-bbbb-bbbb-222222222222", "聊天", "text", "22222222-2222-2222-2222-222222222222", 1},
		{"bbbbbbbb-bbbb-bbbb-bbbb-222222222223", "语音", "voice", "22222222-2222-2222-2222-222222222222", 2},
		
		{"cccccccc-cccc-cccc-cccc-333333333331", "规则", "text", "33333333-3333-3333-3333-333333333333", 0},
		{"cccccccc-cccc-cccc-cccc-333333333332", "聊天", "text", "33333333-3333-3333-3333-333333333333", 1},
		{"cccccccc-cccc-cccc-cccc-333333333333", "语音", "voice", "33333333-3333-3333-3333-333333333333", 2},
	}

	for _, ch := range guildChannels {
		_, err := DB.Exec(`
			INSERT INTO channels (id, name, type, guild_id, position)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, type = EXCLUDED.type
		`, ch.id, ch.name, ch.typ, ch.guildID, ch.position)
		if err != nil {
			log.Printf("Warning: failed to insert channel %s: %v", ch.name, err)
		} else {
			log.Printf("Inserted/updated channel: %s", ch.name)
		}
	}

	log.Println("Default channels seeded successfully")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
