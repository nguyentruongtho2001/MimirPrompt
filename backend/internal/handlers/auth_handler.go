package handlers

import (
	"net/http"
	"os"
	"time"

	"mimirprompt/internal/models"
	"mimirprompt/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key - in production, use environment variable
var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "mimirprompt-secret-key-change-in-production" // Default for development
	}
	return secret
}

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	repo *repository.AuthRepository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(repo *repository.AuthRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

// Login handles user login
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by username
	user, err := h.repo.GetUserByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Verify password
	if !h.repo.VerifyPassword(user.PasswordHash, input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	token, expiresIn, err := generateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := models.AuthResponse{
		Token:     token,
		ExpiresIn: expiresIn,
	}
	response.User.ID = user.ID
	response.User.Username = user.Username

	c.JSON(http.StatusOK, response)
}

// Register handles user registration (admin only in production)
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.repo.CreateUser(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      userID,
		"message": "User registered successfully",
	})
}

// GetProfile returns the current user's profile
// GET /api/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.repo.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

// ChangePassword handles password change
// PUT /api/auth/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify current password
	user, err := h.repo.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !h.repo.VerifyPassword(user.PasswordHash, input.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Update password
	if err := h.repo.UpdatePassword(userID.(int), input.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// generateToken creates a new JWT token
func generateToken(userID int, username string) (string, int64, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
	expiresIn := int64(24 * 60 * 60)                 // 24 hours in seconds

	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)

	return tokenString, expiresIn, err
}

// GetJWTSecret returns the JWT secret for middleware
func GetJWTSecret() []byte {
	return jwtSecret
}
