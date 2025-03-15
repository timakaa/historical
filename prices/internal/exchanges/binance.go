package exchanges

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	pb "github.com/timakaa/historical-common/proto"
)

// BinanceAdapter implements the adapter for Binance exchange
type BinanceAdapter struct {
	client *binance.Client
}

// NewBinanceAdapter creates a new adapter for Binance
func NewBinanceAdapter() *BinanceAdapter {
	return &BinanceAdapter{
		client: binance.NewClient("", ""), // API keys not needed for public endpoints
	}
}

// GetName returns the name of the exchange
func (a *BinanceAdapter) GetName() string {
	return "binance"
}

// GetHistoricalPrices retrieves historical price data from Binance
func (a *BinanceAdapter) GetHistoricalPrices(ctx context.Context, ticker string, limit int64) ([]*pb.PricesResponse, error) {
	log.Printf("Getting historical prices from Binance for %s", ticker)

	// Set default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	// Fetch data from Binance API
	klines, err := a.client.NewKlinesService().
		Symbol(ticker).
		Interval("1d"). // Daily candles
		Limit(int(limit)).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error fetching data from Binance: %v", err)
	}

	// Convert data to response format
	prices := make([]*pb.PricesResponse, 0, len(klines))
	for _, k := range klines {
		// Convert string values to float64
		open, _ := strconv.ParseFloat(k.Open, 64)
		high, _ := strconv.ParseFloat(k.High, 64)
		low, _ := strconv.ParseFloat(k.Low, 64)
		close, _ := strconv.ParseFloat(k.Close, 64)
		volume, _ := strconv.ParseFloat(k.Volume, 64)

		// Convert timestamp to date
		date := time.Unix(k.OpenTime/1000, 0).Format("2006-01-02")

		prices = append(prices, &pb.PricesResponse{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	// Reverse the slice to get the correct chronological order
	for i, j := 0, len(prices)-1; i < j; i, j = i+1, j-1 {
		prices[i], prices[j] = prices[j], prices[i]
	}

	return prices, nil
}
