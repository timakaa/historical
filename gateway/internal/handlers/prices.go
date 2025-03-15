package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/timakaa/historical-common/proto"
)

type PricesHandler struct {
	pricesClient proto.PricesClient
	authClient   proto.AuthClient
}

func NewPricesHandler(pricesClient proto.PricesClient, authClient proto.AuthClient) *PricesHandler {
	return &PricesHandler{
		pricesClient: pricesClient,
		authClient:   authClient,
	}
}

func (h *PricesHandler) HandleGetHistoricalPrices(c *gin.Context) {
	exchange := c.Param("exchange")
	ticker := c.Param("ticker")
	token := c.GetHeader("x-api-key")
	limitStr := c.Query("limit")

	var limit int64
	if limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
		limit = parsedLimit
	}

	// Если токен указан, проверяем количество оставшихся свечей
	if token != "" {
		// Получаем информацию о токене
		tokenInfoReq := &proto.GetTokenInfoRequest{
			Token: token,
		}

		tokenInfo, err := h.authClient.GetTokenInfo(c.Request.Context(), tokenInfoReq)
		if err != nil {
			log.Printf("Error getting token info: %v", err)
			// Можно продолжить выполнение или вернуть ошибку, в зависимости от требований
		} else {
			// Проверяем, достаточно ли оставшихся свечей
			log.Printf("Token %s has %d candles left", token, tokenInfo.CandlesLeft)

			// Если у токена недостаточно свечей, возвращаем ошибку
			if tokenInfo.CandlesLeft <= 0 || tokenInfo.CandlesLeft < limit {
				c.JSON(http.StatusForbidden, gin.H{"error": "no candles left in your token"})
				return
			}
		}
	}

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

	// После получения всех цен, уменьшаем количество оставшихся свечей
	if token != "" {
		updateReq := &proto.UpdateTokenCandlesLeftRequest{
			Token:           token,
			DecreaseCandles: int64(len(prices)),
		}

		_, err := h.authClient.UpdateTokenCandlesLeft(c.Request.Context(), updateReq)
		if err != nil {
			log.Printf("Error updating candles left: %v", err)
		}
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
