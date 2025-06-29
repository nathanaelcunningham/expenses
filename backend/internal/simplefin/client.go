package simplefin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var (
	ErrNoAccessToken       = errors.New("no access token found")
	ErrDecodeAccessToken   = errors.New("error decoding access token")
	ErrRequestFailed       = errors.New("request failed")
	ErrInvalidBaseURL      = errors.New("invalid base URL")
	ErrAccountNotFound     = errors.New("account not found")
	ErrTransactionNotFound = errors.New("transactions not found")
)

type ClientOption func(*Client)

type Client struct {
	BaseURL     *url.URL
	client      *http.Client
	timeout     time.Duration
	maxRetries  int
	enableDebug bool
	logger      *log.Logger
}

// WithTimeout sets the client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithRetries sets the maximum number of retries
func WithRetries(retries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithDebug enables debug logging
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.enableDebug = debug
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *log.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithCustomBaseURL sets a custom base URL directly
func WithCustomBaseURL(baseURLStr string) ClientOption {
	return func(c *Client) {
		if cleanURL, err := cleanAndParseURL(baseURLStr); err == nil {
			c.BaseURL = cleanURL
		}
	}
}

// cleanAndParseURL cleans a URL string and parses it into a url.URL
func cleanAndParseURL(rawURL string) (*url.URL, error) {
	// Remove any control characters, whitespace, and newlines
	cleanedURL := strings.TrimSpace(rawURL)

	// Parse the cleaned URL
	parsedURL, err := url.Parse(cleanedURL)
	if err != nil {
		return nil, err
	}

	return parsedURL, nil
}

// NewClient creates a new SimpleFin client with optional configurations
func NewClient(accessToken string, opts ...ClientOption) (*Client, error) {
	accessURLBytes, err := base64.StdEncoding.DecodeString(accessToken)
	if err != nil {
		return nil, ErrDecodeAccessToken
	}

	// Clean the URL string before parsing
	accessURLStr := strings.TrimSpace(string(accessURLBytes))

	// Parse and validate the base URL
	baseURL, err := url.Parse(accessURLStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidBaseURL, err)
	}

	// Initialize client with default values
	client := &Client{
		BaseURL:     baseURL,
		timeout:     10 * time.Second,
		maxRetries:  3,
		enableDebug: false,
		logger:      log.New(os.Stderr, "[simplefin] ", log.LstdFlags),
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Initialize HTTP client
	client.client = &http.Client{
		Timeout: client.timeout,
	}

	// Validate that we have a proper URL
	if client.BaseURL.Scheme == "" || client.BaseURL.Host == "" {
		return nil, ErrInvalidBaseURL
	}

	// Log the base URL (sanitized) if debug is enabled
	client.log("SimpleFin client initialized with base URL: %s", client.sanitizeURL(client.BaseURL.String()))

	return client, nil
}

// log outputs debug messages if debug mode is enabled
func (c *Client) log(format string, v ...any) {
	if c.enableDebug {
		c.logger.Printf(format, v...)
	}
}

// buildURL properly builds a URL for an API endpoint while preserving authentication
func (c *Client) buildURL(endpoint string) string {
	// Split the endpoint into path and query parts
	endpointPath := endpoint
	queryString := ""

	if idx := strings.Index(endpoint, "?"); idx != -1 {
		endpointPath = endpoint[:idx]
		queryString = endpoint[idx:]
	}

	// Ensure endpoint path starts with no slash for consistent joining
	endpointPath = strings.TrimPrefix(endpointPath, "/")

	// Create a copy of the base URL to avoid modifying the original
	endpointURL := *c.BaseURL

	// Clean and join paths (only for the path part, not the query)
	basePath := strings.TrimSuffix(endpointURL.Path, "/")
	endpointURL.Path = path.Join(basePath, endpointPath)

	// Add the query string back
	resultURL := endpointURL.String() + queryString

	return resultURL
}

// doRequest handles the HTTP request process with retries
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Build the full URL
	fullURL := c.buildURL(endpoint)

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			c.log("Retry attempt %d for %s %s", attempt, method, c.sanitizeURL(fullURL))
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				// Continue with retry
			}
		}

		var req *http.Request
		req, err = http.NewRequestWithContext(ctx, method, fullURL, body)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Add("Accept", "application/json")

		c.log("Making request: %s %s", method, c.sanitizeURL(fullURL))
		resp, err = c.client.Do(req)

		// Don't retry if context was canceled or if we got a valid response
		if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			break
		}

		c.log("Request failed: %v", err)

		// Last attempt, return the error
		if attempt == c.maxRetries {
			return nil, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, err)
		}
	}

	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s - %s", ErrRequestFailed, resp.Status, string(bodyBytes))
	}

	return resp, nil
}

// sanitizeURL removes sensitive information from URL for logging
func (c *Client) sanitizeURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "[unparseable-url]"
	}

	// Redact userinfo for security in logs
	if parsedURL.User != nil {
		parsedURL.User = url.UserPassword("REDACTED", "REDACTED")
	}

	return parsedURL.String()
}

// Accounts retrieves all accounts from the SimpleFin API
func (c *Client) Accounts(ctx context.Context) (AccountsResponse, error) {
	resp, err := c.doRequest(ctx, "GET", "/accounts", nil)
	if err != nil {
		return AccountsResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountsResponse{}, fmt.Errorf("reading response body: %w", err)
	}

	var accountsResponse AccountsResponse
	err = json.Unmarshal(body, &accountsResponse)
	if err != nil {
		return AccountsResponse{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	return accountsResponse, nil
}

// Account retrieves the specified account from the SimpleFin API
func (c *Client) Account(ctx context.Context, account_id string) (AccountResponse, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/accounts?account=%s", account_id), nil)
	if err != nil {
		return AccountResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountResponse{}, fmt.Errorf("reading response body: %w", err)
	}

	var accountsResponse AccountsResponse
	err = json.Unmarshal(body, &accountsResponse)
	if err != nil {
		return AccountResponse{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	for _, account := range accountsResponse.Accounts {
		if account.ID == account_id {
			return AccountResponse{Account: account}, nil
		}
	}

	return AccountResponse{}, ErrAccountNotFound
}

func (c *Client) AccountTransactions(ctx context.Context, params AccountTransactionsRequest) (AccountTransactionsResponse, error) {
	endpoint := fmt.Sprintf("/accounts?account=%s&start-date=%d&end-date=%d", params.AccountID, params.StartDate.Unix(), params.EndDate.Unix())

	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return AccountTransactionsResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountTransactionsResponse{}, fmt.Errorf("reading response body: %w", err)
	}

	var accountsResponse AccountsResponse
	err = json.Unmarshal(body, &accountsResponse)
	if err != nil {
		return AccountTransactionsResponse{}, fmt.Errorf("unmarshalling response: %w", err)
	}

	for _, account := range accountsResponse.Accounts {
		if account.ID == params.AccountID {
			return AccountTransactionsResponse{Transactions: account.Transactions}, nil

		}
	}

	return AccountTransactionsResponse{}, ErrTransactionNotFound
}
