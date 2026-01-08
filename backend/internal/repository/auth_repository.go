package repository

import (
	"database/sql"
	"errors"
	"mimirprompt/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// AuthRepository handles database operations for authentication
type AuthRepository struct {
	db *sql.DB
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// GetUserByUsername returns a user by username
func (r *AuthRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, username, password_hash, created_at 
		FROM admin_users WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID returns a user by ID
func (r *AuthRepository) GetUserByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, username, password_hash, created_at 
		FROM admin_users WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new admin user
func (r *AuthRepository) CreateUser(input models.RegisterInput) (int64, error) {
	// Check if username already exists
	var count int
	r.db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE username = ?", input.Username).Scan(&count)
	if count > 0 {
		return 0, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	result, err := r.db.Exec(`
		INSERT INTO admin_users (username, password_hash) VALUES (?, ?)
	`, input.Username, string(hashedPassword))

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// VerifyPassword checks if the provided password matches the stored hash
func (r *AuthRepository) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// UpdatePassword updates a user's password
func (r *AuthRepository) UpdatePassword(userID int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = r.db.Exec("UPDATE admin_users SET password_hash = ? WHERE id = ?", string(hashedPassword), userID)
	return err
}
