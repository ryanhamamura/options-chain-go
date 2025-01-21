package tasty

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
)

// SessionRequest represents the login request payload
type SessionRequest struct {
    Login        string `json:"login"`
    Password     string `json:"password,omitempty"`
    RememberMe   bool   `json:"remember-me,omitempty"`
    RememberToken string `json:"remember-token,omitempty"`
}

// SessionResponse represents the response from the login endpoint
type SessionResponse struct {
    Data struct {
        User struct {
            Email      string `json:"email"`
            Username   string `json:"username"`
            ExternalID string `json:"external-id"`
        } `json:"user"`
        SessionToken  string `json:"session-token"`
        RememberToken string `json:"remember-token,omitempty"`
    } `json:"data"`
    Context string `json:"context"`
}

// Login authenticates with the Tastytrade API and returns a session token
func (c *Client) Login(ctx context.Context, username, password string) (string, error) {
    // Log the request (without password)
    log.Printf("Attempting login for user '%s' to endpoint: %s/sessions", 
        username, c.config.BaseURL)

    // Create login request
    loginReq := SessionRequest{
        Login:      username,
        Password:   password,
        RememberMe: true, // Request a remember token for future use
    }

    jsonData, err := json.Marshal(loginReq)
    if err != nil {
        return "", fmt.Errorf("marshaling login request: %w", err)
    }

    // Create request
    req, err := http.NewRequestWithContext(ctx, "POST", 
        fmt.Sprintf("%s/sessions", c.config.BaseURL), 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return "", fmt.Errorf("creating login request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    
    // Log request headers
    log.Printf("Request Headers:")
    for key, values := range req.Header {
        log.Printf("  %s: %s", key, values)
    }

    // Execute request
    log.Println("Sending login request...")
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("executing login request: %w", err)
    }
    defer resp.Body.Close()

    // Log response details (without sensitive data)
    log.Printf("Response Status: %s (%d)", resp.Status, resp.StatusCode)
    log.Printf("Response Headers:")
    for key, values := range resp.Header {
        log.Printf("  %s: %s", key, values)
    }

    if resp.StatusCode != http.StatusCreated {
        // Read error response body if available
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
    }

    // Parse response
    var loginResp SessionResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        return "", fmt.Errorf("decoding login response: %w", err)
    }

    // Log successful login (without tokens)
    log.Printf("Successfully logged in as user: %s (%s)", 
        loginResp.Data.User.Username, 
        loginResp.Data.User.Email)

    // Store the remember token if provided
    if loginResp.Data.RememberToken != "" {
        c.mu.Lock()
        c.rememberToken = loginResp.Data.RememberToken
        c.mu.Unlock()
        log.Println("Received and stored remember token")
    }

    return loginResp.Data.SessionToken, nil
}

// LoginWithRememberToken attempts to login using a stored remember token
func (c *Client) LoginWithRememberToken(ctx context.Context, username string) (string, error) {
    c.mu.Lock()
    rememberToken := c.rememberToken
    c.mu.Unlock()

    if rememberToken == "" {
        return "", fmt.Errorf("no remember token available")
    }

    loginReq := SessionRequest{
        Login:         username,
        RememberToken: rememberToken,
        RememberMe:    true, // Request a new remember token
    }

    jsonData, err := json.Marshal(loginReq)
    if err != nil {
        return "", fmt.Errorf("marshaling login request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", 
        fmt.Sprintf("%s/sessions", c.config.BaseURL), 
        bytes.NewBuffer(jsonData))
    if err != nil {
        return "", fmt.Errorf("creating login request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("executing login request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        // Clear the invalid remember token
        c.mu.Lock()
        c.rememberToken = ""
        c.mu.Unlock()
        return "", fmt.Errorf("login failed with status: %d", resp.StatusCode)
    }

    var loginResp SessionResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        return "", fmt.Errorf("decoding login response: %w", err)
    }

    // Store the new remember token
    if loginResp.Data.RememberToken != "" {
        c.mu.Lock()
        c.rememberToken = loginResp.Data.RememberToken
        c.mu.Unlock()
    }

    return loginResp.Data.SessionToken, nil
}

// Logout destroys the current session
func (c *Client) Logout(ctx context.Context) error {
    if c.sessionToken == "" {
        return fmt.Errorf("no active session")
    }

    req, err := http.NewRequestWithContext(ctx, "DELETE", 
        fmt.Sprintf("%s/sessions", c.config.BaseURL), 
        nil)
    if err != nil {
        return fmt.Errorf("creating logout request: %w", err)
    }

    req.Header.Set("Authorization", c.sessionToken)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("executing logout request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent {
        return fmt.Errorf("logout failed with status: %d", resp.StatusCode)
    }

    // Clear session and remember tokens
    c.mu.Lock()
    c.sessionToken = ""
    c.rememberToken = ""
    c.mu.Unlock()

    return nil
}
