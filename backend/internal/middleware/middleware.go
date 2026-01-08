package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"mimirprompt/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GetJWTSecret gets JWT secret from auth_handler to ensure consistency
func getJWTSecret() []byte {
	return handlers.GetJWTSecret()
}

// getAllowedOrigins returns CORS origins from environment or default
func getAllowedOrigins() string {
	origins := os.Getenv("CORS_ORIGINS")
	if origins == "" {
		return "*" // Default for development
	}
	return origins
}

// CORS returns a middleware that handles CORS
func CORS() gin.HandlerFunc {
	allowedOrigins := getAllowedOrigins()
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if allowedOrigins == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Check against comma-separated list of allowed origins
			for _, allowed := range strings.Split(allowedOrigins, ",") {
				if strings.TrimSpace(allowed) == origin {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger returns a middleware that logs requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if status >= 400 {
			gin.DefaultErrorWriter.Write([]byte(
				c.Request.Method + " " + path + " " + latency.String() + "\n",
			))
		}
	}
}

// Recovery returns a middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// AuthRequired returns a middleware that requires JWT authentication
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return getJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user info in context
		userID := int(claims["user_id"].(float64))
		username := claims["username"].(string)

		c.Set("userID", userID)
		c.Set("username", username)

		c.Next()
	}
}

// OptionalAuth returns a middleware that optionally validates JWT
// If token is present and valid, user info is set in context
// If token is absent or invalid, request continues without user info
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return getJWTSecret(), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("userID", int(claims["user_id"].(float64)))
				c.Set("username", claims["username"].(string))
			}
		}

		c.Next()
	}
}
