package config

import "github.com/caarlos0/env/v10"

type Config struct {
	GRPC struct {
		Port string `env:"GRPC_PORT" envDefault:"50057"`
	}

	Clients struct {
		ProviderAddr string `env:"PROVIDER_GRPC_ADDR" envDefault:"127.0.0.1:50058"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
