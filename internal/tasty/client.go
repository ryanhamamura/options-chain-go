package tasty

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/ryanhamamura/options-chain-go/internal/models"
)

const (
    defaultVersion = "0.1-DXF-JS/0.3.0"
    dataFormat    = "COMPACT"
)

// Client handles communication with Tastytrade API and DXLink
type Client struct {
    config     Config
    httpClient *http.Client
    
    // Authentication
    sessionToken  string
    rememberToken string
    
    // DXLink related
    wsConn        *websocket.Conn
    keepaliveDone chan struct{}
    mu            sync.Mutex

    // Connection management
    reconnectManager *reconnectManager
    subscriptions   []DXSubscription
    transformer     *DataTransformer
    
    // Error handling
    errorHandler      func(error)
    disconnectHandler func()
    reconnectHandler  func()
}

// NewClient creates a new Tastytrade API client
func NewClient(config Config) *Client {
    return &Client{
        config: config,
        httpClient: &http.Client{
            Timeout: time.Second * 30,
        },
        keepaliveDone: make(chan struct{}),
        reconnectManager: newReconnectManager(DefaultReconnectConfig),
        transformer: NewDataTransformer(),
        errorHandler: func(err error) {
            log.Printf("DXLink error: %v", err)
        },
        disconnectHandler: func() {
            log.Println("DXLink disconnected")
        },
        reconnectHandler: func() {
            log.Println("DXLink reconnected")
        },
    }
}

// SetSessionToken sets the session token for the client
func (c *Client) SetSessionToken(token string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.sessionToken = token
}

// GetSessionToken gets the current session token
func (c *Client) GetSessionToken() string {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.sessionToken
}

// GetQuoteToken fetches a new API quote token
func (c *Client) GetQuoteToken(ctx context.Context) (*QuoteTokenResponse, error) {
    if c.sessionToken == "" {
        return nil, fmt.Errorf("session token not set")
    }

    req, err := http.NewRequestWithContext(ctx, "GET", c.config.BaseURL+"/api-quote-tokens", nil)
    if err != nil {
        return nil, fmt.Errorf("creating request: %w", err)
    }

    req.Header.Set("Authorization", c.sessionToken)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("executing request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
    }

    var tokenResp QuoteTokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
        return nil, fmt.Errorf("decoding response: %w", err)
    }

    return &tokenResp, nil
}

// ConnectDXLink establishes a WebSocket connection to DXLink
func (c *Client) ConnectDXLink(ctx context.Context) error {
    if c.config.StreamerURL == "" {
        return fmt.Errorf("streamer URL not initialized")
    }

    conn, _, err := websocket.DefaultDialer.Dial(c.config.StreamerURL, nil)
    if err != nil {
        return fmt.Errorf("establishing WebSocket connection: %w", err)
    }
    c.wsConn = conn

    // Send SETUP message
    setup := DXSetupMessage{
        DXMessage: DXMessage{
            Type:    "SETUP",
            Channel: 0,
        },
        Version:               defaultVersion,
        KeepaliveTimeout:     60,
        AcceptKeepaliveTimeout: 60,
    }
    if err := c.writeJSON(setup); err != nil {
        return fmt.Errorf("sending setup message: %w", err)
    }

    // Send AUTH message
    auth := DXAuthMessage{
        DXMessage: DXMessage{
            Type:    "AUTH",
            Channel: 0,
        },
        Token: c.sessionToken,
    }
    if err := c.writeJSON(auth); err != nil {
        return fmt.Errorf("sending auth message: %w", err)
    }

    // Start keepalive routine
    go c.keepaliveRoutine(ctx)

    return nil
}

// Subscribe sets up a channel and subscribes to market data
func (c *Client) Subscribe(ctx context.Context, channel int, subscriptions []DXSubscription) error {
    // Store subscriptions for reconnection
    c.subscriptions = subscriptions

    // Open channel
    channelReq := DXChannelRequest{
        DXMessage: DXMessage{
            Type:    "CHANNEL_REQUEST",
            Channel: channel,
        },
        Service: "FEED",
        Parameters: map[string]string{
            "contract": "AUTO",
        },
    }
    if err := c.writeJSON(channelReq); err != nil {
        return fmt.Errorf("requesting channel: %w", err)
    }

    // Setup feed
    feedSetup := DXFeedSetup{
        DXMessage: DXMessage{
            Type:    "FEED_SETUP",
            Channel: channel,
        },
        AcceptAggregationPeriod: 0.1,
        AcceptDataFormat:        dataFormat,
        AcceptEventFields: map[string][]string{
            "Quote":  {"eventType", "eventSymbol", "bidPrice", "askPrice", "bidSize", "askSize"},
            "Greeks": {"eventType", "eventSymbol", "volatility", "delta", "gamma", "theta", "rho", "vega"},
            "Trade":  {"eventType", "eventSymbol", "price", "dayVolume", "size"},
        },
    }
    if err := c.writeJSON(feedSetup); err != nil {
        return fmt.Errorf("setting up feed: %w", err)
    }

    // Subscribe to market data
    sub := DXFeedSubscription{
        DXMessage: DXMessage{
            Type:    "FEED_SUBSCRIPTION",
            Channel: channel,
        },
        Reset: true,
        Add:   subscriptions,
    }
    if err := c.writeJSON(sub); err != nil {
        return fmt.Errorf("subscribing to feed: %w", err)
    }

    return nil
}

// StartReading starts reading market data events
func (c *Client) StartReading(ctx context.Context, callback func(models.OptionChain)) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                var event MarketDataEvent
                err := c.wsConn.ReadJSON(&event)
                if err != nil {
                    c.errorHandler(fmt.Errorf("reading market data: %w", err))
                    return
                }
                c.transformer.HandleEvent(event)
                chain := c.transformer.GetOptionChain(event.EventSymbol)
                callback(chain)
            }
        }
    }()
}

// keepaliveRoutine sends keepalive messages every 30 seconds
func (c *Client) keepaliveRoutine(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            msg := DXMessage{
                Type:    "KEEPALIVE",
                Channel: 0,
            }
            if err := c.writeJSON(msg); err != nil {
                log.Printf("Error sending keepalive: %v", err)
                return
            }
        }
    }
}

// writeJSON sends a message to the WebSocket connection with mutex protection
func (c *Client) writeJSON(v interface{}) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.wsConn.WriteJSON(v)
}

// Close closes the WebSocket connection
func (c *Client) Close() error {
    if c.wsConn != nil {
        close(c.keepaliveDone)
        return c.wsConn.Close()
    }
    return nil
}

// SetErrorHandler sets the handler for client errors
func (c *Client) SetErrorHandler(handler func(error)) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.errorHandler = handler
}

// SetDisconnectHandler sets the handler for disconnection events
func (c *Client) SetDisconnectHandler(handler func()) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.disconnectHandler = handler
}

// SetReconnectHandler sets the handler for successful reconnections
func (c *Client) SetReconnectHandler(handler func()) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.reconnectHandler = handler
}
