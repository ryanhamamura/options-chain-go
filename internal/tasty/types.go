package tasty 

import "time" 

// QuoteTokenResponse represents the response from /api-quote-tokens
type QuoteTokenResponse struct {
    Data struct {
        Token     string `json:"token"`
        DXLinkURL string `json:"dxlink-url"`
        Level     string `json:"level"`
    } `json:"data"`
    Context string `json:"context"`
}

// DXMessage represents a base DXLink message
type DXMessage struct {
    Type    string `json:"type"`
    Channel int    `json:"channel"`
}

// DXSetupMessage for initial connection setup
type DXSetupMessage struct {
    DXMessage
    Version               string `json:"version"`
    KeepaliveTimeout     int    `json:"keepaliveTimeout"`
    AcceptKeepaliveTimeout int  `json:"acceptKeepaliveTimeout"`
}

// DXAuthMessage for authentication
type DXAuthMessage struct {
    DXMessage
    Token string `json:"token"`
}

// DXChannelRequest for opening a feed channel
type DXChannelRequest struct {
    DXMessage
    Service    string            `json:"service"`
    Parameters map[string]string `json:"parameters"`
}

// DXFeedSetup for configuring the data feed
type DXFeedSetup struct {
    DXMessage
    AcceptAggregationPeriod float64               `json:"acceptAggregationPeriod"`
    AcceptDataFormat        string                `json:"acceptDataFormat"`
    AcceptEventFields       map[string][]string   `json:"acceptEventFields"`
}

// DXSubscription represents a single market data subscription
type DXSubscription struct {
    Type   string `json:"type"`
    Symbol string `json:"symbol"`
}

// DXFeedSubscription for subscribing to market data
type DXFeedSubscription struct {
    DXMessage
    Reset  bool            `json:"reset"`
    Add    []DXSubscription `json:"add,omitempty"`
    Remove []DXSubscription `json:"remove,omitempty"`
}

// MarketDataEvent represents different types of market data
type MarketDataEvent struct {
    EventType   string    `json:"eventType"`
    EventSymbol string    `json:"eventSymbol"`
    Timestamp   time.Time `json:"timestamp"`

    // Quote fields
    BidPrice float64 `json:"bidPrice,omitempty"`
    AskPrice float64 `json:"askPrice,omitempty"`
    BidSize  float64 `json:"bidSize,omitempty"`
    AskSize  float64 `json:"askSize,omitempty"`

    // Greeks fields
    Volatility float64 `json:"volatility,omitempty"`
    Delta      float64 `json:"delta,omitempty"`
    Gamma      float64 `json:"gamma,omitempty"`
    Theta      float64 `json:"theta,omitempty"`
    Rho        float64 `json:"rho,omitempty"`
    Vega       float64 `json:"vega,omitempty"`

    // Trade fields
    Price      float64 `json:"price,omitempty"`
    DayVolume  float64 `json:"dayVolume,omitempty"`
    Size       float64 `json:"size,omitempty"`
}
