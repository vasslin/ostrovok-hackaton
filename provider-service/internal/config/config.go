package config

import "github.com/caarlos0/env/v10"

type Config struct {
	Mode         string `env:"PROVIDER_MODE" envDefault:"demo"`
	RegistryFile string `env:"PROVIDER_CONFIG_FILE" envDefault:"../configs/providers.json"`
	GRPC         struct {
		Port string `env:"GRPC_PORT" envDefault:"50058"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
