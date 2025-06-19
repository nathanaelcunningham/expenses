package turso

import "time"

type DatabaseInfo struct {
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Hostname string    `json:"hostname"`
	DbId     string    `json:"DbId"`
	Version  string    `json:"version"`
	Created  time.Time `json:"created"`
}

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

type SeedDatabaseType string

const (
	SeedDatabaseTypeDatabase       SeedDatabaseType = "database"
	SeedDatabaseTypeDatabaseUpload SeedDatabaseType = "database_upload"
)

type CreateDatabaseRequest struct {
	Name  string                    `json:"name"`
	Group string                    `json:"group"` // Default group is "default"
	Seed  CreateDatabaseRequestSeed `json:"seed"`
}

type CreateDatabaseRequestSeed struct {
	Type SeedDatabaseType `json:"type"`
	Name string           `json:"name"`
}

type LocationsResponse struct {
	Locations map[string]string `json:"locations"`
}
