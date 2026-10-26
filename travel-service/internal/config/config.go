package config

import "github.com/caarlos0/env/v10"

type Config struct {
	GRPC struct {
		Port string `env:"GRPC_PORT" envDefault:"50051"`
	}

	Clients struct {
		SessionAddr string `env:"SESSION_GRPC_ADDR" envDefault:"127.0.0.1:50052"`
		AIAddr      string `env:"AI_GRPC_ADDR" envDefault:"127.0.0.1:50055"`
		MapAddr     string `env:"MAP_GRPC_ADDR" envDefault:"127.0.0.1:50056"`
		RouteAddr   string `env:"ROUTE_GRPC_ADDR" envDefault:"127.0.0.1:50057"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
