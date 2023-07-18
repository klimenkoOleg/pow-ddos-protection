package server

import (
	"crypto/rsa"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"pow-ddos-protection/internal/core/config"
	"pow-ddos-protection/internal/core/encryption"
)

const (
	baseConfigPath = "config/server-config.yaml"
	envConfigPath  = "config/server-config-%s.yaml"
	rsaPrivateKey  = "config/key.pem"
)

type ServerConfig struct {
	AppName            string `yaml:"app-name"`
	Port               string `yaml:"port"`
	HashcashZerosCount int    `yaml:"hashcash-zeros-count"`
	HashcashTimeout    int    `yaml:"hashcash-timeout"`
	PrivateKey         *rsa.PrivateKey
}

// LoadServerConfig loads the configuration from the config/server-config.yaml file.
func LoadServerConfig(log *zap.Logger) (*ServerConfig, error) {
	appCfg := &ServerConfig{}
	err := config.LoadAppConfig(appCfg, baseConfigPath, envConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed load configs")
	}

	privateKey, err := encryption.LoadPrivateRSA(rsaPrivateKey, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed parsing env vars")
	}
	appCfg.PrivateKey = privateKey

	return appCfg, nil
}
