package config

import "github.com/caarlos0/env/v10"

type Config struct {
	GRPC struct {
		Port string `env:"GRPC_PORT" envDefault:"50057"`
	}

	Clients struct {
		ProviderAddr string `env:"PROVIDER_GRPC_ADDR" envDefault:"127.0.0.1:50058"`
		AIAddr       string `env:"AI_GRPC_ADDR" envDefault:"127.0.0.1:50055"`
	}
	Join struct {
		MaxJoins      int `env:"ROUTE_MAX_JOINS" envDefault:"4"`
		RoutesPerJoin int `env:"ROUTE_ROUTES_PER_JOIN" envDefault:"3"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
