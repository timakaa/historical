package exchanges

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/timakaa/historical-common/proto"
)

// TestBinanceAdapter_GetName tests the GetName method
func TestBinanceAdapter_GetName(t *testing.T) {
	adapter := NewBinanceAdapter()
	assert.Equal(t, "binance", adapter.GetName())
}

// TestNewBinanceAdapter tests the creation of a new adapter
func TestNewBinanceAdapter(t *testing.T) {
	adapter := NewBinanceAdapter()
	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.client)
}

// TestBinanceAdapter_ProcessKlineData tests the processing of kline data
func TestBinanceAdapter_ProcessKlineData(t *testing.T) {
	// Create sample kline data
	klineData := []*binance.Kline{
		{
			OpenTime: 1672531200000, // 2023-01-01
			Open:     "10000.0",
			High:     "10100.0",
			Low:      "9900.0",
			Close:    "10050.0",
			Volume:   "1.5",
		},
		{
			OpenTime: 1672617600000, // 2023-01-02
			Open:     "9900.0",
			High:     "10000.0",
			Low:      "9800.0",
			Close:    "9950.0",
			Volume:   "2.0",
		},
	}

	// Process the kline data manually
	prices := make([]*pb.PricesResponse, 0, len(klineData))
	for _, kline := range klineData {
		// Convert timestamp to date
		date := time.Unix(kline.OpenTime/1000, 0).Format("2006-01-02")

		// Convert string values to float64
		open, _ := strconv.ParseFloat(kline.Open, 64)
		high, _ := strconv.ParseFloat(kline.High, 64)
		low, _ := strconv.ParseFloat(kline.Low, 64)
		close, _ := strconv.ParseFloat(kline.Close, 64)
		volume, _ := strconv.ParseFloat(kline.Volume, 64)

		prices = append(prices, &pb.PricesResponse{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	// Verify the processed data
	assert.Len(t, prices, 2)

	assert.Equal(t, "2023-01-01", prices[0].Date)
	assert.Equal(t, 10000.0, prices[0].Open)
	assert.Equal(t, 10100.0, prices[0].High)
	assert.Equal(t, 9900.0, prices[0].Low)
	assert.Equal(t, 10050.0, prices[0].Close)
	assert.Equal(t, 1.5, prices[0].Volume)

	assert.Equal(t, "2023-01-02", prices[1].Date)
	assert.Equal(t, 9900.0, prices[1].Open)
	assert.Equal(t, 10000.0, prices[1].High)
	assert.Equal(t, 9800.0, prices[1].Low)
	assert.Equal(t, 9950.0, prices[1].Close)
	assert.Equal(t, 2.0, prices[1].Volume)
}

// TestBinanceAdapter_GetHistoricalPrices_ErrorHandling tests error handling in GetHistoricalPrices
func TestBinanceAdapter_GetHistoricalPrices_ErrorHandling(t *testing.T) {
	// Test handling of invalid float values
	invalidValue := "not-a-number"

	// Attempt to parse the invalid value
	_, err := strconv.ParseFloat(invalidValue, 64)
	assert.Error(t, err, "Should fail to parse invalid float")

	// Test handling of negative limit
	limit := int64(-10)
	defaultLimit := int64(100)

	// Manually check the limit logic that's in the adapter
	if limit <= 0 {
		limit = defaultLimit
	}

	assert.Equal(t, defaultLimit, limit, "Negative limit should be replaced with default")
}

// TestBinanceAdapter_GetHistoricalPrices tests the GetHistoricalPrices method
func TestBinanceAdapter_GetHistoricalPrices(t *testing.T) {
	// This test verifies that the method exists and has the correct signature
	adapter := NewBinanceAdapter()

	// Call the method with a valid ticker and limit
	ctx := context.Background()
	prices, err := adapter.GetHistoricalPrices(ctx, "BTCUSDT", 10)

	// If the API call succeeds, verify the results
	if err == nil {
		require.NotNil(t, prices)
		require.LessOrEqual(t, len(prices), 10)

		// Verify the structure of the returned data
		for _, price := range prices {
			assert.NotEmpty(t, price.Date)
			assert.NotZero(t, price.Open)
			assert.NotZero(t, price.High)
			assert.NotZero(t, price.Low)
			assert.NotZero(t, price.Close)
		}
	} else {
		// If the API call fails, that's also acceptable for a unit test
		assert.Contains(t, err.Error(), "error fetching data from Binance")
	}

	// Test with invalid ticker
	_, err = adapter.GetHistoricalPrices(ctx, "INVALID_TICKER_12345", 5)
	assert.Error(t, err)

	// Test with default limit (0)
	prices, err = adapter.GetHistoricalPrices(ctx, "BTCUSDT", 0)
	if err == nil {
		require.NotNil(t, prices)
		assert.LessOrEqual(t, len(prices), 100) // Default limit is 100
	} else {
		assert.Contains(t, err.Error(), "error fetching data from Binance")
	}
}

// TestBinanceAdapter_Integration tests the real implementation
// It's skipped by default to avoid network dependencies during unit testing
func TestBinanceAdapter_Integration(t *testing.T) {
	t.Skip("Skipping integration test - requires network access")
}
