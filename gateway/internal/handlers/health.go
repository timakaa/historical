package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timakaa/historical-gateway/internal/services"
)

type HealthHandler struct {
	historicalService services.HistoricalService
	authService       services.AuthService
}

func NewHealthHandler(
	historicalService services.HistoricalService,
	authService services.AuthService,
) *HealthHandler {
	return &HealthHandler{
		historicalService: historicalService,
		authService:       authService,
	}
}

func (h *HealthHandler) Handle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"services": gin.H{
			"historical": h.historicalService.CheckHealth(),
			"access":     h.authService.CheckHealth(),
		},
	})
}
