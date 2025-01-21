package tasty

import (
    "sync"
    "time"

    "github.com/ryanhamamura/options-chain-go/internal/models"
)

// DataTransformer aggregates and transforms DXLink market data
type DataTransformer struct {
    mu sync.RWMutex
    // Cache for latest market data by symbol
    quotes  map[string]MarketDataEvent
    greeks  map[string]MarketDataEvent
    trades  map[string]MarketDataEvent
    summary map[string]MarketDataEvent
}

func NewDataTransformer() *DataTransformer {
    return &DataTransformer{
        quotes:  make(map[string]MarketDataEvent),
        greeks:  make(map[string]MarketDataEvent),
        trades:  make(map[string]MarketDataEvent),
        summary: make(map[string]MarketDataEvent),
    }
}

// HandleEvent processes incoming market data events
func (t *DataTransformer) HandleEvent(event MarketDataEvent) {
    t.mu.Lock()
    defer t.mu.Unlock()

    switch event.EventType {
    case "Quote":
        t.quotes[event.EventSymbol] = event
    case "Greeks":
        t.greeks[event.EventSymbol] = event
    case "Trade":
        t.trades[event.EventSymbol] = event
    case "Summary":
        t.summary[event.EventSymbol] = event
    }
}

// GetOptionChain generates an option chain from the latest market data
func (t *DataTransformer) GetOptionChain(symbol string) models.OptionChain {
    t.mu.RLock()
    defer t.mu.RUnlock()

    // Get latest data for the symbol
    quote := t.quotes[symbol]
    greeks := t.greeks[symbol]
    trade := t.trades[symbol]

    // Transform into OptionData
    optionData := models.OptionData{
        Strike:     parseStrikeFromSymbol(symbol),  // You'll need to implement this
        Expiration: parseExpirationFromSymbol(symbol), // And this
        Type:       parseOptionType(symbol),        // And this
        Bid:        quote.BidPrice,
        Ask:        quote.AskPrice,
        LastPrice:  trade.Price,
        Volume:     int(trade.DayVolume),
        OpenInt:    0, // Get from Summary event if available
        Delta:      greeks.Delta,
        Gamma:      greeks.Gamma,
        Theta:      greeks.Theta,
        Vega:       greeks.Vega,
        ImpliedVol: greeks.Volatility,
    }

    // Create the chain
    chain := models.OptionChain{
        Symbol:     symbol,
        Underlying: trade.Price,
        Updated:    time.Now(),
    }

    // Sort into calls and puts
    if optionData.Type == "call" {
        chain.Calls = append(chain.Calls, optionData)
    } else {
        chain.Puts = append(chain.Puts, optionData)
    }

    return chain
}

// Helper functions for parsing option symbols
// You'll need to implement these based on Tastytrade's symbol format
func parseStrikeFromSymbol(symbol string) float64 {
    // Implementation depends on symbol format
    return 0
}

func parseExpirationFromSymbol(symbol string) string {
    // Implementation depends on symbol format
    return ""
}

func parseOptionType(symbol string) string {
    // Implementation depends on symbol format
    return ""
}
