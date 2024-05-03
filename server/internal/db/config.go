package db

import (
	"fmt"
	"os"
)

// Config specifies the configurations to connect to the database.
type Config struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Database        string `yaml:"database"`
	PasswordEnvName string `yaml:"passwordEnvName"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("port must be greater than 0")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database is required")
	}
	if c.PasswordEnvName == "" {
		return fmt.Errorf("passwordEnvName is required")
	}
	return nil
}

// password returns the password for the connection to the database.
func (c Config) password() string {
	return os.Getenv(c.PasswordEnvName)
}
