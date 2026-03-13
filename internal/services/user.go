package services

import (
	"golang.org/x/crypto/bcrypt"
	"discord-backend/internal/database"
	"discord-backend/internal/models"

	"github.com/google/uuid"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Create(username, email, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	discriminator := "0001"
	var user models.User

	err = database.DB.QueryRow(`
		INSERT INTO users (username, email, password_hash, discriminator)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, avatar, discriminator, created_at, updated_at
	`, username, email, string(hash), discriminator).Scan(
		&user.ID, &user.Username, &user.Email, &user.Avatar,
		&user.Discriminator, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, username, email, avatar, discriminator, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Username, &user.Email, &user.Avatar,
		&user.Discriminator, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, username, email, password_hash, avatar, discriminator, created_at, updated_at
		FROM users WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Avatar, &user.Discriminator, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, username, email, password_hash, avatar, discriminator, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Avatar, &user.Discriminator, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

func (s *UserService) UpdateAvatar(userID uuid.UUID, avatar string) error {
	_, err := database.DB.Exec(`
		UPDATE users SET avatar = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
	`, avatar, userID)
	return err
}
