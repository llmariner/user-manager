package config

import (
	"fmt"
	"os"

	"github.com/llmariner/api-usage/pkg/sender"
	"github.com/llmariner/common/pkg/db"
	"gopkg.in/yaml.v3"
)

// DefaultOrganizationConfig is the default organization configuration.
type DefaultOrganizationConfig struct {
	Title    string   `yaml:"title"`
	UserIDs  []string `yaml:"userIds"`
	TenantID string   `yaml:"tenantId"`
}

// validate validates the configuration.
func (c *DefaultOrganizationConfig) validate() error {
	if c.Title == "" {
		return fmt.Errorf("title must be set")
	}
	if len(c.UserIDs) == 0 {
		return fmt.Errorf("userIds must be set")
	}
	if c.TenantID == "" {
		return fmt.Errorf("tenantId must be set")
	}
	return nil
}

// DefaultProjectConfig is the default project configuration.
type DefaultProjectConfig struct {
	Title               string `yaml:"title"`
	KubernetesNamespace string `yaml:"kubernetesNamespace"`
}

// validate validates the configuration.
func (c *DefaultProjectConfig) validate() error {
	if c.Title == "" {
		return fmt.Errorf("title must be set")
	}
	return nil
}

// DefaultAPIKeyConfig is the default API key configuration.
type DefaultAPIKeyConfig struct {
	Name             string `yaml:"name"`
	Secret           string `yaml:"secret"`
	UserID           string `yaml:"userId"`
	IsServiceAccount bool   `yaml:"isServiceAccount"`
}

// validate validates the configuration.
func (c *DefaultAPIKeyConfig) validate() error {
	if c.Name == "" {
		return fmt.Errorf("name must be set")
	}
	if c.Secret == "" {
		return fmt.Errorf("secret must be set")
	}
	if c.IsServiceAccount {
		if c.UserID != "" {
			return fmt.Errorf("userId must not be set")
		}
	} else {
		if c.UserID == "" {
			return fmt.Errorf("userId must be set")
		}
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

// AssumeRoleConfig is the assume role configuration.
type AssumeRoleConfig struct {
	RoleARN    string `yaml:"roleArn"`
	ExternalID string `yaml:"externalId"`
}

func (c *AssumeRoleConfig) validate() error {
	if c.RoleARN == "" {
		return fmt.Errorf("roleArn must be set")
	}
	return nil
}

// KMSConfig is AWS KMS configuration.
type KMSConfig struct {
	Enable     bool              `yaml:"enable"`
	KeyAlias   string            `yaml:"keyAlias"`
	Region     string            `yaml:"region"`
	AssumeRole *AssumeRoleConfig `yaml:"assumeRole"`
}

// validate validates the configuration.
func (c *KMSConfig) validate() error {
	if !c.Enable {
		return nil
	}
	if c.KeyAlias == "" {
		return fmt.Errorf("keyAlias must be set")
	}
	if c.Region == "" {
		return fmt.Errorf("region must be set")
	}
	if ar := c.AssumeRole; ar != nil {
		if err := ar.validate(); err != nil {
			return fmt.Errorf("assumeRole: %s", err)
		}
	}
	return nil
}

// Config is the configuration.
type Config struct {
	GRPCPort         int `yaml:"grpcPort"`
	HTTPPort         int `yaml:"httpPort"`
	InternalGRPCPort int `yaml:"internalGrpcPort"`

	Database    db.Config     `yaml:"database"`
	AuthConfig  AuthConfig    `yaml:"auth"`
	UsageSender sender.Config `yaml:"usageSender"`

	DefaultOrganization *DefaultOrganizationConfig `yaml:"defaultOrganization"`
	DefaultProject      *DefaultProjectConfig      `yaml:"defaultProject"`
	DefaultAPIKeys      []DefaultAPIKeyConfig      `yaml:"defaultApiKeys"`

	KMSConfig KMSConfig `yaml:"kms"`

	Debug DebugConfig `yaml:"debug"`
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

	if oc := c.DefaultOrganization; oc != nil {
		if err := oc.validate(); err != nil {
			return err
		}
	}
	if dp := c.DefaultProject; dp != nil {
		if c.DefaultOrganization == nil {
			return fmt.Errorf("defaultOrganization must be set when defaultProject is set")
		}
		if err := c.DefaultProject.validate(); err != nil {
			return err
		}
	}
	if len(c.DefaultAPIKeys) > 0 {
		if c.DefaultOrganization == nil {
			return fmt.Errorf("defaultOrganization must be set when defaultApiKeys is set")
		}
		if c.DefaultProject == nil {
			return fmt.Errorf("defaultProject must be set when defaultApiKeys is set")
		}

		for _, k := range c.DefaultAPIKeys {
			if err := k.validate(); err != nil {
				return fmt.Errorf("defaultApiKey: %s", err)
			}
		}
	}

	if err := c.AuthConfig.validate(); err != nil {
		return err
	}
	if err := c.UsageSender.Validate(); err != nil {
		return err
	}
	if err := c.KMSConfig.validate(); err != nil {
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
