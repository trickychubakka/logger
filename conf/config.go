// Package conf -- пакет, содержащий объекты конфигураций logger.
package conf

import "crypto/rsa"

// Config конфигурация logger сервера.
type Config struct {
	Database struct {
		Host     string `mapstructure:"host"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Dbname   string `mapstructure:"dbname"`
		Sslmode  string `mapstructure:"sslmode"`
	}
}

// AgentConfig конфигурация logger агента.
type AgentConfig struct {
	PollInterval     int            // Agent metric polling interval.
	ReportInterval   int            // Agent metric report interval.
	Address          string         // Logger server address and port.
	Logfile          string         // Agent log file.
	Key              string         // HMAC key.
	RateLimit        int            // Rate limit for agent connections to server.
	PProfHTTPEnabled bool           // Flag for enabling pprof web server.
	PathToPublicKey  string         // Path to RSA public key file.
	PublicKey        *rsa.PublicKey // RSA public key.
}
