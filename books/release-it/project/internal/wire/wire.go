// Package wire holds the JSON shapes exchanged between litefleet services and
// with tenants. Gateway, control and node all import it so the formats agree.
package wire

// Statement is one SQL statement with positional arguments.
type Statement struct {
	SQL  string `json:"sql"`
	Args []any  `json:"args,omitempty"`
}

// QueryRequest is the body of POST .../query (tenant-facing and node-facing).
type QueryRequest struct {
	SQL  string `json:"sql"`
	Args []any  `json:"args,omitempty"`
}

// QueryResult is the body of a successful query. Truncated is set when the
// answer hit the platform's cap (1 000 rows or 1 MB).
type QueryResult struct {
	Columns   []string `json:"columns"`
	Rows      [][]any  `json:"rows"`
	Truncated bool     `json:"truncated"`
}

// ExecuteRequest is the body of POST .../execute: every statement runs in one
// transaction.
type ExecuteRequest struct {
	Statements []Statement `json:"statements"`
}

// ExecuteResult is the body of a successful execute.
type ExecuteResult struct {
	RowsAffected    int64 `json:"rows_affected"`
	LastInsertRowID int64 `json:"last_insert_rowid"`
}

// Error is the typed error every service returns.
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorBody wraps Error on the wire: {"error": {...}}.
type ErrorBody struct {
	Error Error `json:"error"`
}

// ResolveRequest asks control to authenticate a key and locate a database.
type ResolveRequest struct {
	APIKey string `json:"api_key"`
	DBName string `json:"db_name"`
}

// Route says where a tenant's database lives.
type Route struct {
	TenantID   string `json:"tenant_id"`
	DatabaseID string `json:"database_id"`
	NodeURL    string `json:"node_url"`
	Node       string `json:"node"`
	Epoch      int    `json:"epoch"`
}

// Database is a tenant's database as the registry knows it.
type Database struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Node       string `json:"node"`
	State      string `json:"state,omitempty"`
	BytesTotal int64  `json:"bytes_total"`
	RowsTotal  int64  `json:"rows_total"`
	SizeBytes  int64  `json:"size_bytes"`
	LastBackup string `json:"last_backup,omitempty"`
}

// CreateDatabaseRequest is the body of POST /v1/databases.
type CreateDatabaseRequest struct {
	Name string `json:"name"`
}

// ProvisionRequest is control's command to a node to create a database file.
type ProvisionRequest struct {
	DatabaseID string `json:"database_id"`
	TenantID   string `json:"tenant_id"`
	Epoch      int    `json:"epoch"`
}

// Heartbeat is what a node reports to control every few seconds.
type Heartbeat struct {
	Node      string           `json:"node"`
	URL       string           `json:"url"`
	FDs       int              `json:"fds"`
	DiskFree  int64            `json:"disk_free"`
	Databases []DatabaseReport `json:"databases"`
}

// DatabaseReport is one database's size and its activity since the previous
// heartbeat.
type DatabaseReport struct {
	ID         string `json:"id"`
	Bytes      int64  `json:"bytes"`
	WALBytes   int64  `json:"wal_bytes"`
	Reads      int64  `json:"reads"`
	ReadRows   int64  `json:"read_rows"`
	ReadBytes  int64  `json:"read_bytes"`
	Writes     int64  `json:"writes"`
	WriteRows  int64  `json:"write_rows"`
	WriteBytes int64  `json:"write_bytes"`
	LastBackup string `json:"last_backup,omitempty"`
}

// UsageEvent is one billable statement, sent from a node to metering.
type UsageEvent struct {
	EventID    string `json:"event_id"`
	TenantID   string `json:"tenant_id"`
	DatabaseID string `json:"database_id"`
	Kind       string `json:"kind"`
	Rows       int64  `json:"rows"`
	Bytes      int64  `json:"bytes"`
	At         string `json:"at"`
}

// TenantUsage is one row of the fleet-wide usage report.
type TenantUsage struct {
	TenantID     string `json:"tenant_id"`
	Statements   int64  `json:"statements"`
	Rows         int64  `json:"rows"`
	Bytes        int64  `json:"bytes"`
	StorageBytes int64  `json:"storage_bytes"`
}

// UsageReport is the body of GET /v1/admin/usage.
type UsageReport struct {
	Window  string        `json:"window"`
	Tenants []TenantUsage `json:"tenants"`
}
