package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Address     string `yaml:"address"`
	OpenAPIPath string `yaml:"openapiPath"`
	SwaggerPath string `yaml:"swaggerPath"`
}

type JWTConfig struct {
	TokenHeader string `yaml:"tokenHeader"`
	TokenPrefix string `yaml:"tokenPrefix"`
	TokenSecret string `yaml:"tokenSecret"`
}

type SecureConfig struct {
	Ignore     []string `yaml:"ignore"`
	PublicKey  string   `yaml:"publicKey"`
	PrivateKey string   `yaml:"privateKey"`
}

type ShopSuiteConfig struct {
	Version        string `yaml:"version"`
	AppDebug       bool   `yaml:"appDebug"`
	DbDebug        bool   `yaml:"dbDebug"`
	UploadDir      string `yaml:"uploadDir"`
	CacheEnable    bool   `yaml:"cacheEnable"`
	CacheNamespace string `yaml:"cacheNamespace"`
	CacheSeparator string `yaml:"cacheSeparator"`
	UrlBase        string `yaml:"urlBase"`
	UrlH5          string `yaml:"urlH5"`
	UrlPc          string `yaml:"urlPc"`
}

// LicenceConfig 与 golershop manifest/config 中 licence 段一致（授权密文）。
type LicenceConfig struct {
	Key string `yaml:"key"`
}

type AppConfig struct {
	Server    ServerConfig    `yaml:"server"`
	JWT       JWTConfig       `yaml:"jwt"`
	Secure    SecureConfig    `yaml:"secure"`
	ShopSuite ShopSuiteConfig `yaml:"shopSuite"`
	Licence   LicenceConfig   `yaml:"licence"`
	Database  DatabaseConfig  `yaml:"database"`
}

type DatabaseLinkConfig struct {
	Link string `yaml:"link"`
}

type DatabaseConfig struct {
	Groups map[string]DatabaseLinkConfig `yaml:",inline"`
}

func (d DatabaseConfig) Link(group string) string {
	if d.Groups == nil {
		return ""
	}
	if cfg, ok := d.Groups[group]; ok {
		return cfg.Link
	}
	return ""
}

func Load(path string) (*AppConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg AppConfig
	if err = yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Server.Address == "" {
		cfg.Server.Address = ":8000"
	}
	if cfg.JWT.TokenHeader == "" {
		cfg.JWT.TokenHeader = "Authorization"
	}
	if cfg.JWT.TokenPrefix == "" {
		cfg.JWT.TokenPrefix = "Bearer "
	}
	if cfg.ShopSuite.CacheSeparator == "" {
		cfg.ShopSuite.CacheSeparator = ":"
	}
	return &cfg, nil
}
