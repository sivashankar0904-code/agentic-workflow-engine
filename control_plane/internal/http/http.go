package http

import "time"

// Config holds HTTP server listen and timeout settings.
type Config struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Addr returns the "host:port" listen address.
func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}
