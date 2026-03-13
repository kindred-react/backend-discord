package services

import (
	"crypto/sha256"

	"discord-backend/config"
	"discord-backend/internal/models"

	"github.com/google/uuid"
)

func GetBuiltInUserByEmail(email string) *models.User {
	cfg := config.Load()

	emailToID := map[string]string{
		"jia@test.com":  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"yi@test.com":   "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		"bing@test.com": "cccccccc-cccc-cccc-cccc-cccccccccccc",
	}

	for _, bu := range cfg.BuiltInUsers {
		if bu.Email == email {
			userIDStr := emailToID[email]
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				hash := sha256.Sum256([]byte(bu.Email))
				var newID uuid.UUID
				copy(newID[:], hash[:16])
				userID = newID
			}
			return &models.User{
				ID:            userID,
				Username:      bu.Username,
				Email:         bu.Email,
				PasswordHash:  "$builtin$" + bu.Password,
				Discriminator: "0000",
			}
		}
	}
	return nil
}

func ValidateBuiltInPassword(user *models.User, password string) bool {
	if len(user.PasswordHash) > 11 && user.PasswordHash[:11] == "$builtin$" {
		return user.PasswordHash[11:] == password
	}
	return false
}

func GetBuiltInUserByUsername(username string) *models.User {
	cfg := config.Load()

	usernameToID := map[string]string{
		"甲":  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"乙": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		"丙": "cccccccc-cccc-cccc-cccc-cccccccccccc",
	}

	for _, bu := range cfg.BuiltInUsers {
		if bu.Username == username {
			userIDStr := usernameToID[username]
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				hash := sha256.Sum256([]byte(bu.Email))
				var newID uuid.UUID
				copy(newID[:], hash[:16])
				userID = newID
			}
			return &models.User{
				ID:            userID,
				Username:      bu.Username,
				Email:         bu.Email,
				PasswordHash:  "$builtin$" + bu.Password,
				Discriminator: "0000",
			}
		}
	}
	return nil
}
