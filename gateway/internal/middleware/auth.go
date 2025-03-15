package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timakaa/historical-common/proto"
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Create gRPC request
		req := &proto.ValidateRequest{
			Token: token,
		}

		// Call gRPC service
		resp, err := m.authClient.ValidateToken(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("userID", resp.UserId)
		c.Set("permissions", resp.Permissions)
		c.Next()
	}
}
