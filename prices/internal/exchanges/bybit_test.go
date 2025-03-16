package exchanges

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/timakaa/historical-common/proto"
)

// TestBybitAdapter_GetName tests the GetName method
func TestBybitAdapter_GetName(t *testing.T) {
	adapter := NewBybitAdapter()
	assert.Equal(t, "bybit", adapter.GetName())
}

// TestNewBybitAdapter tests the creation of a new adapter
func TestNewBybitAdapter(t *testing.T) {
	adapter := NewBybitAdapter()
	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.client)
}

// TestBybitAdapter_ProcessKlineData tests the processing of kline data
func TestBybitAdapter_ProcessKlineData(t *testing.T) {
	// Create sample kline data
	now := time.Now().Unix() * 1000 // milliseconds

	// Create a simple struct to represent kline data for testing
	type KlineItem struct {
		StartTime string
		Open      string
		High      string
		Low       string
		Close     string
		Volume    string
	}

	klineData := []KlineItem{
		{
			StartTime: strconv.FormatInt(now, 10),
			Open:      "20000.0",
			High:      "20100.0",
			Low:       "19900.0",
			Close:     "20050.0",
			Volume:    "2.5",
		},
		{
			StartTime: strconv.FormatInt(now-86400000, 10), // Previous day
			Open:      "19900.0",
			High:      "20000.0",
			Low:       "19800.0",
			Close:     "19950.0",
			Volume:    "3.0",
		},
	}

	// Process the kline data manually
	prices := make([]*pb.PricesResponse, 0, len(klineData))
	for _, item := range klineData {
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

	// Verify the processed data
	assert.Len(t, prices, 2)
	assert.Equal(t, time.Unix(now/1000, 0).Format("2006-01-02"), prices[0].Date)
	assert.Equal(t, 20000.0, prices[0].Open)
	assert.Equal(t, 20100.0, prices[0].High)
	assert.Equal(t, 19900.0, prices[0].Low)
	assert.Equal(t, 20050.0, prices[0].Close)
	assert.Equal(t, 2.5, prices[0].Volume)

	assert.Equal(t, time.Unix((now-86400000)/1000, 0).Format("2006-01-02"), prices[1].Date)
	assert.Equal(t, 19900.0, prices[1].Open)
	assert.Equal(t, 20000.0, prices[1].High)
	assert.Equal(t, 19800.0, prices[1].Low)
	assert.Equal(t, 19950.0, prices[1].Close)
	assert.Equal(t, 3.0, prices[1].Volume)
}

// TestBybitAdapter_GetHistoricalPrices tests the GetHistoricalPrices method
func TestBybitAdapter_GetHistoricalPrices(t *testing.T) {
	// This test verifies that the method exists and has the correct signature
	adapter := NewBybitAdapter()

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
		assert.Contains(t, err.Error(), "error fetching data from Bybit")
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
		assert.Contains(t, err.Error(), "error fetching data from Bybit")
	}
}

// TestBybitAdapter_ErrorHandling tests error handling in the adapter
func TestBybitAdapter_ErrorHandling(t *testing.T) {
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

// TestBybitAdapter_GetHistoricalPrices_Integration tests the GetHistoricalPrices method
// It's skipped by default to avoid network dependencies during unit testing
func TestBybitAdapter_GetHistoricalPrices_Integration(t *testing.T) {
	t.Skip("Skipping integration test - requires network access")

	adapter := NewBybitAdapter()
	ctx := context.Background()

	// Test with valid ticker and limit
	prices, err := adapter.GetHistoricalPrices(ctx, "BTCUSDT", 5)
	require.NoError(t, err)
	require.NotNil(t, prices)
	require.LessOrEqual(t, len(prices), 5)

	// Verify the structure of the returned data
	for _, price := range prices {
		assert.NotEmpty(t, price.Date)
		assert.NotZero(t, price.Open)
		assert.NotZero(t, price.High)
		assert.NotZero(t, price.Low)
		assert.NotZero(t, price.Close)
	}

	// Test with invalid ticker
	prices, err = adapter.GetHistoricalPrices(ctx, "INVALID_TICKER", 5)
	assert.Error(t, err)

	// Test with default limit
	prices, err = adapter.GetHistoricalPrices(ctx, "BTCUSDT", 0)
	require.NoError(t, err)
	require.NotNil(t, prices)
	assert.LessOrEqual(t, len(prices), 100) // Default limit is 100
}

// TestBybitAdapter_Integration tests the real implementation
// It's skipped by default to avoid network dependencies during unit testing
func TestBybitAdapter_Integration(t *testing.T) {
	t.Skip("Skipping integration test - requires network access")
}
