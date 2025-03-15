package exchanges

import (
	"context"

	pb "github.com/timakaa/historical-common/proto"
)

// ExchangeAdapter defines the interface for all exchange adapters
type ExchangeAdapter interface {
	// GetName returns the name of the exchange
	GetName() string

	// GetHistoricalPrices retrieves historical price data for the specified ticker
	GetHistoricalPrices(ctx context.Context, ticker string, limit int64) ([]*pb.PricesResponse, error)
}

// ExchangeFactory is a factory for creating exchange adapters
type ExchangeFactory struct {
	adapters map[string]ExchangeAdapter
}

// NewExchangeFactory creates a new factory with registered adapters
func NewExchangeFactory() *ExchangeFactory {
	factory := &ExchangeFactory{
		adapters: make(map[string]ExchangeAdapter),
	}

	// Register adapters for supported exchanges
	factory.RegisterAdapter(NewBinanceAdapter())
	factory.RegisterAdapter(NewBybitAdapter())

	return factory
}

// RegisterAdapter registers a new adapter in the factory
func (f *ExchangeFactory) RegisterAdapter(adapter ExchangeAdapter) {
	f.adapters[adapter.GetName()] = adapter
}

// GetAdapter returns an adapter for the specified exchange
func (f *ExchangeFactory) GetAdapter(exchange string) (ExchangeAdapter, bool) {
	adapter, exists := f.adapters[exchange]
	return adapter, exists
}
