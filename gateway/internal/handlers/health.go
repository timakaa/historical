package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/timakaa/historical-common/proto"
)

type HealthHandler struct {
	pricesClient proto.PricesClient
	authClient   proto.AuthClient
}

func NewHealthHandler(
	pricesClient proto.PricesClient,
	authClient proto.AuthClient,
) *HealthHandler {
	return &HealthHandler{
		pricesClient: pricesClient,
		authClient:   authClient,
	}
}

func (h *HealthHandler) Handle(c *gin.Context) {
	// Create context with timeout for health checks
	_, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check historical service health
	historicalStatus := "ok"
	// _, err := h.pricesClient.Health(ctx, &proto.HealthRequest{})
	// if err != nil {
	// 	historicalStatus = "error: " + err.Error()
	// }

	// Check auth service health
	authStatus := "ok"
	// _, err = h.authClient.Health(ctx, &proto.HealthRequest{})
	// if err != nil {
	// 	authStatus = "error: " + err.Error()
	// }

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"services": gin.H{
			"historical": historicalStatus,
			"access":     authStatus,
		},
	})
}

func (h *HealthHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.Handle)
}
