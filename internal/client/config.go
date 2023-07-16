package client

import (
	"crypto/rsa"
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
	AppName               string `yaml:"app-name"`
	ServerURL             string `yaml:"server-url"`
	HashcashMaxIterations int    `yaml:"hashcash-max-iterations"`
	PrivateKey            *rsa.PrivateKey
}

// LoadServerConfig loads the configuration from the config/server-config.yaml file.
func LoadClientConfig(log *zap.Logger) (*ClientConfig, error) {
	appCfg := &ClientConfig{}
	err := config.LoadAppConfig(appCfg, baseConfigPath, envConfigPath)
	if err != nil {
		return nil, config.ErrRsaFile.Wrap(err)
	}

	privateKey, err := encryption.LoadPrivateRSA(rsaPrivateKey, log)
	if err != nil {
		return nil, config.ErrRsaFile.Wrap(err)
	}
	appCfg.PrivateKey = privateKey

	return appCfg, nil
}
