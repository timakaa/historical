package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timakaa/historical-common/proto"
)

type AuthHandler struct {
	authClient proto.AuthClient
}

func NewAuthHandler(authClient proto.AuthClient) *AuthHandler {
	return &AuthHandler{
		authClient: authClient,
	}
}

func (h *AuthHandler) HandleCreateToken(c *gin.Context) {
	req := &proto.CreateTokenRequest{
		Permissions: []string{"read:prices"},
		ExpiresIn:   60 * 60 * 24 * 90, // 90 days in seconds
	}

	resp, err := h.authClient.CreateToken(c.Request.Context(), req)
	if err != nil {
		fmt.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": resp.Token})
}

func (h *AuthHandler) HandleValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization header"})
		return
	}

	req := &proto.ValidateRequest{
		Token: token,
	}

	resp, err := h.authClient.ValidateToken(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":      resp.UserId,
		"permissions": resp.Permissions,
	})
}

func (h *AuthHandler) HandleRevokeToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization header"})
		return
	}

	req := &proto.RevokeTokenRequest{
		Token: token,
	}

	_, err := h.authClient.RevokeToken(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token revoked successfully"})
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/token", h.HandleCreateToken)
		authGroup.GET("/validate", h.HandleValidateToken)
		authGroup.DELETE("/token", h.HandleRevokeToken)
	}
}
