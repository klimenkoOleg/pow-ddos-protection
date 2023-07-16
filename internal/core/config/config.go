package config

import (
	"crypto/rsa"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"pow-ddos-protection/internal/core/encryption"
	"pow-ddos-protection/internal/core/errors"
)

const (
	// ErrValidation is returned when the configuration is invalid.
	ErrValidation = errors.Error("invalid configuration")
	// ErrEnvVars is returned when the environment variables are invalid.
	ErrEnvVars = errors.Error("failed parsing env vars")
	// ErrRead is returned when the configuration file cannot be read.
	ErrRead = errors.Error("failed to read file")
	// ErrUnmarshal is returned when the configuration file cannot be unmarshalled.
	ErrUnmarshal = errors.Error("failed to unmarshal file")
	// ErrRsaFile is returned when RSA privae key file reading failed to read.
	ErrRsaFile = errors.Error("failed parsing env vars")
)

var (
	baseConfigPath = "config/server-config.yaml"
	envConfigPath  = "config/server-config-%s.yaml"
	rsaPrivateKey  = "config/key.pem"
)

// Config represents the configuration of our application.
type Config struct {
	Log        *zap.Logger
	AppConfig  *AppConfig
	PrivateKey *rsa.PrivateKey
}

type AppConfig struct {
	AppName               string `yaml:"app-name"`
	Port                  string `yaml:"port"`
	ServerURL             string `yaml:"server-url"`
	HashcashZerosCount    int    `yaml:"hashcash-zeros-count"`
	HashcashTimeout       int    `yaml:"hashcash-timeout"`
	HashcashMaxIterations int    `yaml:"hashcash-max-iterations"`
}

// Load loads the configuration from the config/server-config.yaml file.
func LoadAppConfig(log *zap.Logger) (*Config, error) {
	privateKey, err := encryption.LoadPrivateRSA(rsaPrivateKey, *log)
	if err != nil {
		return nil, ErrRsaFile.Wrap(err)
	}

	cfg := &Config{Log: log, PrivateKey: privateKey}

	if err := cfg.loadFromFiles(); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, ErrEnvVars.Wrap(err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, ErrValidation.Wrap(err)
	}

	return cfg, nil
}

func (ac *Config) loadFromFiles() error {
	// load base config
	if err := ac.loadYaml(baseConfigPath); err != nil {
		return err
	}

	// load environment-specific config
	environ := os.Getenv("OKLIMENKO_ENVIRONMENT")
	// no need to load env-specific yaml if no ENV variable is set
	if environ == "" {
		return nil
	}

	p := fmt.Sprintf(envConfigPath, environ)
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		if err := ac.loadYaml(p); err != nil {
			return err
		}
	}

	return nil
}

func (ac *Config) loadYaml(filename string) error {
	ac.Log.Info("Loading configuration", zap.String("path", filename))

	data, err := os.ReadFile(filename)
	if err != nil {
		return ErrRead.Wrap(err)
	}

	appConfig := &AppConfig{}
	if err := yaml.Unmarshal(data, appConfig); err != nil {
		return ErrUnmarshal.Wrap(err)
	}
	ac.AppConfig = appConfig

	return nil
}
