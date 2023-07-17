package config

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

func LoadAppConfig(appConfig interface{}, baseConfigPath, envConfigPath string) error {
	if err := loadFromFiles(appConfig, baseConfigPath, envConfigPath); err != nil {
		return err
	}

	if err := env.Parse(appConfig); err != nil {
		return errors.Wrap(err, "failed parsing env vars")
	}

	validate := validator.New()
	if err := validate.Struct(appConfig); err != nil {
		return errors.Wrap(err, "invalid configuration")
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
		return errors.Wrap(err, "failed to read file")
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return errors.Wrap(err, "failed to unmarshal file")
	}

	return nil
}
