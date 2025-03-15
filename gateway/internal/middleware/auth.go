package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/timakaa/historical-common/proto"
)

const (
	UserIDKey      = "userID"
	PermissionsKey = "permissions"
)

type AuthMiddleware struct {
	authClient proto.AuthClient
}

func NewAuthMiddleware(authClient proto.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("x-api-key")
		if token == "" {
			log.Println("Missing authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		// Create gRPC request
		req := &proto.ValidateRequest{
			Token: token,
		}

		// Add timeout to context
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Call gRPC service
		resp, err := m.authClient.ValidateToken(ctx, req)
		if err != nil {
			log.Printf("Error validating token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Check if token is valid
		if !resp.IsValid {
			log.Printf("Invalid token: %s", token)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token is invalid or expired"})
			return
		}

		log.Printf("User authenticated: %s with permissions: %v", resp.UserId, resp.Permissions)

		// Store user info in context for later use
		c.Set(UserIDKey, resp.UserId)
		c.Set(PermissionsKey, resp.Permissions)

		c.Next()
	}
}

// Helper functions to get user data from context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	return userID.(string)
}

func GetPermissions(c *gin.Context) []string {
	permissions, exists := c.Get(PermissionsKey)
	if !exists {
		return nil
	}
	return permissions.([]string)
}
