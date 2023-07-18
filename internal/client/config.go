package client

import (
	"crypto/rsa"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"pow-ddos-protection/internal/core/config"
	"pow-ddos-protection/internal/core/encryption"
)

const (
	baseConfigPath = "config/client-config.yaml"
	envConfigPath  = "config/client-config-%s.yaml"
	rsaPrivateKey  = "config/key.pem"
)

type ClientConfig struct {
	AppName                 string `yaml:"app-name"`
	ServerAddress           string `yaml:"server-address" env:"SERVER_ADDRESS"`
	RequestsCreationTimeout int    `yaml:"requests-creation-timeout"`
	NumberOfClients         int    `yaml:"number-of-clients"`
	RequestsPerClient       int    `yaml:"requests-per-client"`
	HashcashMaxIterations   int    `yaml:"hashcash-max-iterations"`
	PrivateKey              *rsa.PrivateKey
}

// LoadClientConfig loads the configuration from the config/server-config.yaml file.
func LoadClientConfig(log *zap.Logger) (*ClientConfig, error) {
	appCfg := &ClientConfig{}
	err := config.LoadAppConfig(appCfg, baseConfigPath, envConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed load configs")
	}

	privateKey, err := encryption.LoadPrivateRSA(rsaPrivateKey, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed loD RSA file")
	}
	appCfg.PrivateKey = privateKey

	return appCfg, nil
}
