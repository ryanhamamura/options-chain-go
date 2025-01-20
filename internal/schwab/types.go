package schwab

import "time"

// Credentials holds API authentication details
type Credentials struct {
    APIKey    string
    APISecret string
    // Add any other required authentication fields
}

// OptionQuote represents the quote data from Schwab API
type OptionQuote struct {
    ContractID   string    `json:"contractId"`
    Strike       float64   `json:"strike"`
    Expiration   time.Time `json:"expiration"`
    Type         string    `json:"type"`
    Bid          float64   `json:"bid"`
    Ask          float64   `json:"ask"`
    Last         float64   `json:"last"`
    Volume       int       `json:"volume"`
    OpenInterest int       `json:"openInterest"`
    Greeks       Greeks    `json:"greeks"`
}

// Greeks represents option Greeks from Schwab API
type Greeks struct {
    Delta float64 `json:"delta"`
    Gamma float64 `json:"gamma"`
    Theta float64 `json:"theta"`
    Vega  float64 `json:"vega"`
    Rho   float64 `json:"rho"`
}

// OptionsChainResponse represents the API response for options chain data
type OptionsChainResponse struct {
    Symbol          string        `json:"symbol"`
    UnderlyingPrice float64      `json:"underlyingPrice"`
    Timestamp       time.Time     `json:"timestamp"`
    Contracts       []OptionQuote `json:"contracts"`
    ExpirationDates []time.Time   `json:"expirationDates"`
    Strikes         []float64     `json:"strikes"`
}

// StreamUpdate represents a real-time update from the streaming API
type StreamUpdate struct {
    Type      string      `json:"type"`
    Timestamp time.Time   `json:"timestamp"`
    Data      OptionQuote `json:"data"`
}
