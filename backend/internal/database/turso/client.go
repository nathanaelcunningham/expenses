package turso

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"expenses-backend/internal/logger"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

// Client represents a Turso client with connection management
type Client struct {
	authToken    string
	organization string
	logger       logger.Logger
	connections  sync.Map // map[string]*sql.DB - cached connections
	mu           sync.RWMutex
	maxRetries   int
	retryDelay   time.Duration
	httpClient   *http.Client
	baseURL      string
}

// Config holds Turso client configuration
type Config struct {
	AuthToken    string        `json:"auth_token"`
	Organization string        `json:"organization"`
	MaxRetries   int           `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`
}

// DatabaseInfo represents database metadata
type DatabaseInfo struct {
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Hostname string    `json:"hostname"`
	DbId     string    `json:"DbId"`
	Version  string    `json:"version"`
	Created  time.Time `json:"created"`
}

// TursoAPIResponse represents responses from Turso API
type TursoAPIResponse struct {
	Database *TursoDatabase `json:"database,omitempty"`
	Error    *TursoAPIError `json:"error,omitempty"`
}

type TursoDatabase struct {
	Name      string   `json:"name"`
	Hostname  string   `json:"hostname"`
	DbId      string   `json:"DbId"`
	Locations []string `json:"locations"`
}

type TursoAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreateDatabaseRequest represents the request to create a database
type CreateDatabaseRequest struct {
	Name     string `json:"name"`
	Location string `json:"location,omitempty"`
	Image    string `json:"image,omitempty"`
}

// LocationsResponse represents the available locations
type LocationsResponse struct {
	Locations map[string]string `json:"locations"`
}

// NewClient creates a new Turso client
func NewClient(config Config, log logger.Logger) *Client {
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}

	return &Client{
		authToken:    config.AuthToken,
		organization: config.Organization,
		logger:       log.With(logger.Str("component", "turso-client")),
		maxRetries:   config.MaxRetries,
		retryDelay:   config.RetryDelay,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.turso.tech/v1",
	}
}

// CreateDatabase creates a new Turso database via API
func (c *Client) CreateDatabase(ctx context.Context, name string, location string) (*DatabaseInfo, error) {
	if location == "" {
		location = "ord" // Default to Chicago
	}

	c.logger.Info("Creating Turso database", logger.Str("database", name), logger.Str("location", location))

	// Prepare request
	reqBody := CreateDatabaseRequest{
		Name:     name,
		Location: location,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/organizations/%s/databases", c.baseURL, c.organization)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute request with retries
	var resp *http.Response
	for attempt := range c.maxRetries {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry client errors)
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt < c.maxRetries-1 {
			c.logger.Warn("Database creation attempt failed, retrying",
				err,
				logger.Int("attempt", attempt+1),
				logger.Int("status_code", func() int {
					if resp != nil {
						return resp.StatusCode
					}
					return 0
				}()))

			time.Sleep(c.retryDelay * time.Duration(attempt+1))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create database after %d attempts: %w", c.maxRetries, err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiError TursoAPIResponse
		if err := json.Unmarshal(body, &apiError); err == nil && apiError.Error != nil {
			return nil, fmt.Errorf("turso API error: %s - %s", apiError.Error.Code, apiError.Error.Message)
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse TursoAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResponse.Database == nil {
		return nil, fmt.Errorf("no database info in response")
	}

	// Construct database URL
	dbURL := fmt.Sprintf("libsql://%s.turso.io", apiResponse.Database.Hostname)

	dbInfo := &DatabaseInfo{
		Name:     apiResponse.Database.Name,
		URL:      dbURL,
		Hostname: apiResponse.Database.Hostname,
		DbId:     apiResponse.Database.DbId,
		Version:  "1.0.0",
		Created:  time.Now(),
	}

	c.logger.Info("Database created successfully",
		logger.Str("database", name),
		logger.Str("hostname", apiResponse.Database.Hostname),
		logger.Str("url", dbURL),
		logger.Str("db_id", apiResponse.Database.DbId))

	return dbInfo, nil
}

// DeleteDatabase deletes a Turso database via API
func (c *Client) DeleteDatabase(ctx context.Context, name string) error {
	c.logger.Info("Deleting Turso database", logger.Str("database", name))

	// Create HTTP request
	url := fmt.Sprintf("%s/organizations/%s/databases/%s", c.baseURL, c.organization, name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	// Execute request with retries
	var resp *http.Response
	for attempt := range c.maxRetries {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry client errors)
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt < c.maxRetries-1 {
			c.logger.Warn("Database deletion attempt failed, retrying", err, logger.Int("attempt", attempt+1))

			time.Sleep(c.retryDelay * time.Duration(attempt+1))
		}
	}

	if err != nil {
		return fmt.Errorf("failed to delete database after %d attempts: %w", c.maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var apiError TursoAPIResponse
		if err := json.Unmarshal(body, &apiError); err == nil && apiError.Error != nil {
			return fmt.Errorf("turso API error: %s - %s", apiError.Error.Code, apiError.Error.Message)
		}
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Info("Database deleted successfully", logger.Str("database", name))

	return nil
}

// Connect establishes a connection to a Turso database
func (c *Client) Connect(ctx context.Context, dbURL string) (*sql.DB, error) {
	// Check if we already have a cached connection
	if conn, ok := c.connections.Load(dbURL); ok {
		if db, ok := conn.(*sql.DB); ok {
			// Verify the connection is still alive
			if err := db.PingContext(ctx); err == nil {
				return db, nil
			}
			// Connection is dead, remove from cache
			c.connections.Delete(dbURL)
		}
	}

	c.logger.Debug("Establishing new database connection", logger.Str("url", dbURL))

	// Create connection with retry logic
	var db *sql.DB
	var err error

	for attempt := range c.maxRetries {
		db, err = c.connectWithRetry(ctx, dbURL)
		if err == nil {
			break
		}

		c.logger.Warn("Connection attempt failed, retrying", err, logger.Int("attempt", attempt+1), logger.Int("max_retries", c.maxRetries))

		if attempt < c.maxRetries-1 {
			time.Sleep(c.retryDelay * time.Duration(attempt+1))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect after %d attempts: %w", c.maxRetries, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 30)

	// Cache the connection
	c.connections.Store(dbURL, db)

	c.logger.Info("Database connection established", logger.Str("url", dbURL))
	return db, nil
}

// connectWithRetry performs a single connection attempt
func (c *Client) connectWithRetry(ctx context.Context, dbURL string) (*sql.DB, error) {
	connector, err := libsql.NewConnector(dbURL, libsql.WithAuthToken(c.authToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	db := sql.OpenDB(connector)

	// Test the connection
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// GetConnection retrieves a cached connection or creates a new one
func (c *Client) GetConnection(ctx context.Context, dbURL string) (*sql.DB, error) {
	return c.Connect(ctx, dbURL)
}

// CloseConnection closes and removes a cached connection
func (c *Client) CloseConnection(dbURL string) error {
	if conn, ok := c.connections.LoadAndDelete(dbURL); ok {
		if db, ok := conn.(*sql.DB); ok {
			c.logger.Debug("Closing database connection", logger.Str("url", dbURL))
			return db.Close()
		}
	}
	return nil
}

// CloseAll closes all cached connections
func (c *Client) CloseAll() error {
	var errors []error

	c.connections.Range(func(key, value any) bool {
		if db, ok := value.(*sql.DB); ok {
			if err := db.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close connection %s: %w", key, err))
			}
		}
		c.connections.Delete(key)
		return true
	})

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %v", errors)
	}

	c.logger.Info("All database connections closed")
	return nil
}

// HealthCheck performs a health check on all cached connections
func (c *Client) HealthCheck(ctx context.Context) map[string]error {
	results := make(map[string]error)

	c.connections.Range(func(key, value any) bool {
		dbURL := key.(string)
		if db, ok := value.(*sql.DB); ok {
			err := db.PingContext(ctx)
			results[dbURL] = err
			if err != nil {
				c.logger.Warn("Database health check failed", err, logger.Str("url", dbURL))
			}
		}
		return true
	})

	return results
}

// GetStats returns connection statistics
func (c *Client) GetStats() map[string]sql.DBStats {
	stats := make(map[string]sql.DBStats)

	c.connections.Range(func(key, value any) bool {
		dbURL := key.(string)
		if db, ok := value.(*sql.DB); ok {
			stats[dbURL] = db.Stats()
		}
		return true
	})

	return stats
}
