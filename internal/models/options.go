package models

import "time"

// OptionData represents a single option contract
type OptionData struct {
    Strike      float64 `json:"strike"`
    Expiration  string  `json:"expiration"`
    Type        string  `json:"type"` // "call" or "put"
    Bid         float64 `json:"bid"`
    Ask         float64 `json:"ask"`
    LastPrice   float64 `json:"lastPrice"`
    Volume      int     `json:"volume"`
    OpenInt     int     `json:"openInterest"`
    Delta       float64 `json:"delta"`
    Gamma       float64 `json:"gamma"`
    Theta       float64 `json:"theta"`
    Vega        float64 `json:"vega"`
    ImpliedVol  float64 `json:"impliedVolatility"`
}

// OptionChain represents the full options chain
type OptionChain struct {
    Symbol      string       `json:"symbol"`
    Underlying  float64      `json:"underlyingPrice"`
    Updated     time.Time    `json:"lastUpdated"`
    Calls       []OptionData `json:"calls"`
    Puts        []OptionData `json:"puts"`
}
