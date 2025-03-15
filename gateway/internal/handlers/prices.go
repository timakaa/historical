package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/timakaa/historical-common/proto"
)

type PricesHandler struct {
	pricesClient proto.PricesClient
}

func NewPricesHandler(pricesClient proto.PricesClient) *PricesHandler {
	return &PricesHandler{
		pricesClient: pricesClient,
	}
}

func (h *PricesHandler) HandleGetHistoricalPrices(c *gin.Context) {
	exchange := c.Param("exchange")
	ticker := c.Param("ticker")

	// Create gRPC request
	req := &proto.PricesRequest{
		Exchange: exchange,
		Ticker:   ticker,
	}

	// Call gRPC service
	stream, err := h.pricesClient.GetPrices(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prices"})
		return
	}

	type Price struct {
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}

	// Collect all prices in an array
	var prices []Price
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error receiving price: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error receiving prices"})
			return
		}

		// Add price to array
		prices = append(prices, Price{
			Open:   resp.Open,
			High:   resp.High,
			Low:    resp.Low,
			Close:  resp.Close,
			Volume: resp.Volume,
		})
	}

	// Return all prices as JSON array
	c.JSON(http.StatusOK, gin.H{"prices": prices})
}

func (h *PricesHandler) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	pricesGroup := router.Group("/prices")

	if len(middlewares) > 0 {
		pricesGroup.Use(middlewares...)
	}

	pricesGroup.GET("/:exchange/:ticker", h.HandleGetHistoricalPrices)
}
