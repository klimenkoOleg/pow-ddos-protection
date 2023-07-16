package config

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"os"
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

//// Config represents the configuration of our application.
//type Config struct {
//	appConfig interface{}
//	Log       *zap.Logger
//}

func LoadAppConfig(appConfig interface{}, baseConfigPath, envConfigPath string) error {
	if err := loadFromFiles(appConfig, baseConfigPath, envConfigPath); err != nil {
		return err
	}

	if err := env.Parse(appConfig); err != nil {
		return ErrEnvVars.Wrap(err)
	}

	validate := validator.New()
	if err := validate.Struct(appConfig); err != nil {
		return ErrValidation.Wrap(err)
	}

	return nil
}

func loadFromFiles(appConfig interface{}, baseConfigPath, envConfigPath string) error {
	// load base config
	if err := loadYaml(baseConfigPath, appConfig); err != nil {
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
		if err := loadYaml(p, appConfig); err != nil {
			return err
		}
	}

	return nil
}

func loadYaml(filename string, cfg interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ErrRead.Wrap(err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return ErrUnmarshal.Wrap(err)
	}

	return nil
}
