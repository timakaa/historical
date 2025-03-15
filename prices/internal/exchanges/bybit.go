package exchanges

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hirokisan/bybit/v2"
	pb "github.com/timakaa/historical-common/proto"
)

// BybitAdapter implements the adapter for Bybit exchange
type BybitAdapter struct {
	client *bybit.Client
}

// NewBybitAdapter creates a new adapter for Bybit
func NewBybitAdapter() *BybitAdapter {
	client := bybit.NewClient()
	return &BybitAdapter{
		client: client,
	}
}

// GetName returns the name of the exchange
func (a *BybitAdapter) GetName() string {
	return "bybit"
}

// GetHistoricalPrices retrieves historical price data from Bybit
func (a *BybitAdapter) GetHistoricalPrices(ctx context.Context, ticker string, limit int64) ([]*pb.PricesResponse, error) {
	log.Printf("Getting historical prices from Bybit for %s", ticker)

	// Set default limit if not specified
	limitInt := int(limit)
	if limit <= 0 {
		limitInt = 100
	}

	// Fetch data from Bybit API
	resp, err := a.client.V5().Market().GetKline(bybit.V5GetKlineParam{
		Category: bybit.CategoryV5Spot,
		Symbol:   bybit.SymbolV5(ticker),
		Interval: bybit.Interval("D"), // Daily candles
		Limit:    &limitInt,
	})

	if err != nil {
		return nil, fmt.Errorf("error fetching data from Bybit: %v", err)
	}

	// Check if request was successful
	if resp.RetCode != 0 {
		return nil, fmt.Errorf("Bybit API error: %s", resp.RetMsg)
	}

	// Convert data to response format
	prices := make([]*pb.PricesResponse, 0, len(resp.Result.List))
	for _, item := range resp.Result.List {
		// Convert string values to float64
		open, _ := strconv.ParseFloat(item.Open, 64)
		high, _ := strconv.ParseFloat(item.High, 64)
		low, _ := strconv.ParseFloat(item.Low, 64)
		close, _ := strconv.ParseFloat(item.Close, 64)
		volume, _ := strconv.ParseFloat(item.Volume, 64)

		// Convert timestamp to date
		timestamp, _ := strconv.ParseInt(item.StartTime, 10, 64)
		date := time.Unix(timestamp/1000, 0).Format("2006-01-02")

		prices = append(prices, &pb.PricesResponse{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return prices, nil
}
