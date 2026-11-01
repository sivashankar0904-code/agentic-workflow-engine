package datastore

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Driver    string `json:"driver,omitempty"`
	Host      string `json:"host,omitempty"`
	Port      string `json:"port,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	StoreName string `json:"storeName,omitempty"`
	Schema    string `json:"schema,omitempty"`
	PoolSize  int    `json:"poolSize,omitempty"`
	SSLMode   string `json:"sslMode,omitempty"`

	UseTLS     bool   `json:"useTLS,omitempty"`
	TLSVersion string `json:"tlsVersion,omitempty"`

	DialTimeoutSecs  time.Duration `json:"dialTimeoutSecs,omitempty"`
	ReadTimeoutSecs  time.Duration `json:"readTimeoutSecs,omitempty"`
	WriteTimeoutSecs time.Duration `json:"writeTimeoutSecs,omitempty"`
	IdleTimeoutSecs  time.Duration `json:"idleTimeoutSecs,omitempty"`

	// mongodb specific
	// RetryWrites This parameter enables automatic retryable writes for certain write operations. If a write operation fails due to a network error or a primary election, the MongoDB driver will automatically retry the write operation once. This helps in improving the reliability of the application without having to handle these retries manually in the application code.
	RetryWrites     bool   `json:"retryWrites,omitempty"`
	WriteOperations string `json:"writeOperations,omitempty"`
	AppName         string `json:"appName,omitempty"`
}

// PQDriver wraps pgx.Pool to enable instrumentation
type PQDriver struct {
	pool *pgxpool.Pool
}

// NewPQDriver wraps an established pool.
func NewPQDriver(pool *pgxpool.Pool) *PQDriver {
	return &PQDriver{pool: pool}
}

// Connect builds a connection pool from dscfg and wraps it in a PQDriver.
func Connect(dscfg *Config) (*PQDriver, error) {
	poolcfg, err := pgxpool.ParseConfig(dscfg.ConnURL())
	if err != nil {
		return nil, err
	}

	poolcfg.MaxConnLifetime = dscfg.IdleTimeoutSecs
	poolcfg.MaxConns = int32(dscfg.PoolSize)

	dialer := &net.Dialer{KeepAlive: dscfg.IdleTimeoutSecs * 2}
	dialer.Timeout = dscfg.DialTimeoutSecs
	poolcfg.ConnConfig.DialFunc = dialer.DialContext

	pool, err := pgxpool.NewWithConfig(context.Background(), poolcfg)
	if err != nil {
		return nil, err
	}

	return NewPQDriver(pool), nil
}

func (cfg *Config) ConnURL() string {
	sslMode := strings.TrimSpace(cfg.SSLMode)
	if sslMode == "" {
		sslMode = "disable"
	}

	url := fmt.Sprintf(
		"%s://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Driver,
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.StoreName,
		sslMode,
	)

	if cfg.Schema != "" {
		switch cfg.Driver {
		case "postgres":
			{
				url = fmt.Sprintf("%s&search_path=%s", url, cfg.Schema)
			}
		}
	}

	return url
}
