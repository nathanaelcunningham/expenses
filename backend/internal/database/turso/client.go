package turso

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tursodatabase/libsql-client-go/libsql"
)

type Client struct {
	authToken    string
	apiToken     string
	organization string
	maxRetries   int
	retryDelay   time.Duration
	httpClient   *http.Client
	baseURL      string
}

// Config holds Turso client configuration
type Config struct {
	AuthToken    string        `json:"auth_token"`
	ApiToken     string        `json:"api_token"`
	Organization string        `json:"organization"`
	MaxRetries   int           `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`
}

func NewClient(config Config) *Client {
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = time.Second
	}

	return &Client{
		authToken:    config.AuthToken,
		apiToken:     config.ApiToken,
		organization: config.Organization,
		maxRetries:   config.MaxRetries,
		retryDelay:   config.RetryDelay,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.turso.tech/v1",
	}
}

func (c *Client) CreateDatabase(ctx context.Context, name string, seed string) (*DatabaseInfo, error) {
	reqBody := CreateDatabaseRequest{
		Name:  name,
		Group: "default",
		Seed: CreateDatabaseRequestSeed{
			Type: "databases",
			Name: seed,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/organizations/%s/databases", c.baseURL, c.organization)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := range c.maxRetries {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break // Success or client error (don't retry client errors)
		}

		if resp != nil {
			resp.Body.Close()
		}

		if attempt < c.maxRetries-1 {
			time.Sleep(c.retryDelay * time.Duration(attempt+1))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create database after %d attempts: %w", c.maxRetries, err)
	}
	defer resp.Body.Close()

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

	var apiResponse TursoAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResponse.Database == nil {
		return nil, fmt.Errorf("no database info in response")
	}

	dbURL := fmt.Sprintf("libsql://%s.turso.io", apiResponse.Database.Hostname)

	dbInfo := &DatabaseInfo{
		Name:     apiResponse.Database.Name,
		URL:      dbURL,
		Hostname: apiResponse.Database.Hostname,
		DbId:     apiResponse.Database.DbId,
		Version:  "1.0.0",
		Created:  time.Now(),
	}

	return dbInfo, nil
}

func (c *Client) DeleteDatabase(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/organizations/%s/databases/%s", c.baseURL, c.organization, name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.authToken)

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

	return nil
}

func (c *Client) Connect(ctx context.Context, dbURL string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for attempt := range c.maxRetries {
		db, err = c.connectWithRetry(ctx, dbURL)
		if err == nil {
			break
		}

		if attempt < c.maxRetries-1 {
			time.Sleep(c.retryDelay * time.Duration(attempt+1))
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect after %d attempts: %w", c.maxRetries, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 30)

	return db, nil
}

func (c *Client) connectWithRetry(ctx context.Context, dbURL string) (*sql.DB, error) {
	connector, err := libsql.NewConnector(dbURL, libsql.WithAuthToken(c.authToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	db := sql.OpenDB(connector)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func (c *Client) GetConnection(ctx context.Context, dbURL string) (*sql.DB, error) {
	return c.Connect(ctx, dbURL)
}
