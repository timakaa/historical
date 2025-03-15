package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timakaa/historical-gateway/internal/services"
)

type PricesHandler struct {
	historicalService services.HistoricalService
}

func NewPricesHandler(historicalService services.HistoricalService) *PricesHandler {
	return &PricesHandler{
		historicalService: historicalService,
	}
}

func (h *PricesHandler) HandleGetHistoricalPrices(c *gin.Context) {
	exchange := c.Param("exchange")
	ticker := c.Param("ticker")

	// Get stream from historical service
	stream, err := h.historicalService.GetHistoricalPrices(c.Request.Context(), exchange, ticker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prices"})
		return
	}

	// Set up SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Stream prices to client
	for {
		price, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error receiving price: %v", err)
			break
		}

		// Send price as SSE
		c.SSEvent("price", price)
		c.Writer.Flush()
	}
}
