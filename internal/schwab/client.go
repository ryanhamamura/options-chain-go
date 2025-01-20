package schwab

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"

    "github.com/gorilla/websocket"
)

// Client interfaces with the Schwab API
type Client struct {
    baseURL     string
    wsURL       string
    httpClient  *http.Client
    creds       Credentials
    rateLimiter *RateLimiter
}

// NewClient creates a new Schwab API client
func NewClient(baseURL string, wsURL string, creds Credentials) *Client {
    return &Client{
        baseURL: baseURL,
        wsURL:   wsURL,
        httpClient: &http.Client{
            Timeout: time.Second * 30,
        },
        creds: creds,
        rateLimiter: NewRateLimiter(time.Second / 10), // 10 requests per second
    }
}

// GetOptionsChain fetches the full options chain for a symbol
func (c *Client) GetOptionsChain(ctx context.Context, symbol string) (*OptionsChainResponse, error) {
    endpoint := fmt.Sprintf("%s/v1/markets/options/%s", c.baseURL, url.PathEscape(symbol))
    
    req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }
    
    // Add authentication headers
    c.addAuthHeaders(req)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("executing request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
    }
    
    var chain OptionsChainResponse
    if err := json.NewDecoder(resp.Body).Decode(&chain); err != nil {
        return nil, fmt.Errorf("decoding response: %w", err)
    }
    
    return &chain, nil
}

// StreamOptionQuotes establishes a WebSocket connection for real-time quotes
func (c *Client) StreamOptionQuotes(ctx context.Context, symbol string, callback func(StreamUpdate)) error {
    // Create WebSocket connection
    conn, _, err := websocket.DefaultDialer.Dial(c.wsURL, nil)
    if err != nil {
        return fmt.Errorf("establishing WebSocket connection: %w", err)
    }
    defer conn.Close()

    // Subscribe to updates
    subscribe := map[string]interface{}{
        "type":   "subscribe",
        "symbol": symbol,
    }
    if err := conn.WriteJSON(subscribe); err != nil {
        return fmt.Errorf("subscribing to updates: %w", err)
    }

    // Handle incoming messages
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                var update StreamUpdate
                if err := conn.ReadJSON(&update); err != nil {
                    // Handle error or reconnect
                    continue
                }
                callback(update)
            }
        }
    }()

    return nil
}

// addAuthHeaders adds authentication headers to the request
func (c *Client) addAuthHeaders(req *http.Request) {
    // Add your authentication headers here
    // This will depend on Schwab's authentication requirements
    req.Header.Set("X-API-Key", c.creds.APIKey)
    // Add any other required headers
}

// Helper method to convert Schwab's option quote to our internal model
func convertToOptionData(quote OptionQuote) models.OptionData {
    return models.OptionData{
        Strike:     quote.Strike,
        Expiration: quote.Expiration.Format("2006-01-02"),
        Type:       quote.Type,
        Bid:        quote.Bid,
        Ask:        quote.Ask,
        LastPrice:  quote.Last,
        Volume:     quote.Volume,
        OpenInt:    quote.OpenInterest,
        Delta:      quote.Greeks.Delta,
        Gamma:      quote.Greeks.Gamma,
        Theta:      quote.Greeks.Theta,
        Vega:       quote.Greeks.Vega,
        ImpliedVol: 0, // You might need to calculate this or get it from the API
    }
}
