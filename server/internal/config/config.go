package config

import (
	"fmt"
	"os"

	"github.com/llm-operator/common/pkg/db"
	"gopkg.in/yaml.v3"
)

// DefaultOrganizationConfig is the default organization configuration.
type DefaultOrganizationConfig struct {
	Title   string   `yaml:"title"`
	UserIDs []string `yaml:"userIds"`
}

// validate validates the configuration.
func (c *DefaultOrganizationConfig) validate() error {
	if c.Title == "" {
		return fmt.Errorf("title must be set")
	}
	if len(c.UserIDs) == 0 {
		return fmt.Errorf("userIds must be set")
	}
	return nil
}

// DefaultProjectConfig is the default project configuration.
type DefaultProjectConfig struct {
	Title               string   `yaml:"title"`
	KubernetesNamespace string   `yaml:"kubernetesNamespace"`
	UserIDs             []string `yaml:"userIds"`
}

// validate validates the configuration.
func (c *DefaultProjectConfig) validate() error {
	if c.Title == "" {
		return fmt.Errorf("title must be set")
	}
	if len(c.UserIDs) == 0 {
		return fmt.Errorf("userIds must be set")
	}
	return nil
}

// DebugConfig is the debug configuration.
type DebugConfig struct {
	Standalone bool   `yaml:"standalone"`
	SqlitePath string `yaml:"sqlitePath"`
}

// AuthConfig is the authentication configuration.
type AuthConfig struct {
	Enable                 bool   `yaml:"enable"`
	RBACInternalServerAddr string `yaml:"rbacInternalServerAddr"`
}

// validate validates the configuration.
func (c *AuthConfig) validate() error {
	if !c.Enable {
		return nil
	}
	if c.RBACInternalServerAddr == "" {
		return fmt.Errorf("rbacInternalServerAddr must be set")
	}
	return nil
}

// Config is the configuration.
type Config struct {
	GRPCPort         int `yaml:"grpcPort"`
	HTTPPort         int `yaml:"httpPort"`
	InternalGRPCPort int `yaml:"internalGrpcPort"`

	Database db.Config `yaml:"database"`

	DefaultOrganization DefaultOrganizationConfig `yaml:"defaultOrganization"`
	DefaultProject      DefaultProjectConfig      `yaml:"defaultProject"`

	Debug DebugConfig `yaml:"debug"`

	AuthConfig AuthConfig `yaml:"auth"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.GRPCPort <= 0 {
		return fmt.Errorf("grpcPort must be greater than 0")
	}
	if c.HTTPPort <= 0 {
		return fmt.Errorf("httpPort must be greater than 0")
	}
	if c.InternalGRPCPort <= 0 {
		return fmt.Errorf("internalGrpcPort must be greater than 0")
	}

	if c.Debug.Standalone {
		if c.Debug.SqlitePath == "" {
			return fmt.Errorf("sqlite path must be set")
		}
	} else {
		if err := c.Database.Validate(); err != nil {
			return fmt.Errorf("database: %s", err)
		}
	}

	if err := c.DefaultOrganization.validate(); err != nil {
		return err
	}
	if err := c.DefaultProject.validate(); err != nil {
		return err
	}

	if err := c.AuthConfig.validate(); err != nil {
		return err
	}
	return nil
}

// Parse parses the configuration file at the given path, returning a new
// Config struct.
func Parse(path string) (Config, error) {
	var config Config

	b, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("config: read: %s", err)
	}

	if err = yaml.Unmarshal(b, &config); err != nil {
		return config, fmt.Errorf("config: unmarshal: %s", err)
	}
	return config, nil
}
