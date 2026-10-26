package config

import "github.com/caarlos0/env/v10"

type Config struct {
	HTTP struct {
		Port string `env:"HTTP_PORT" envDefault:"8080"`
	}

	Clients struct {
		TravelAddr  string `env:"TRAVEL_GRPC_ADDR" envDefault:"127.0.0.1:50051"`
		SessionAddr string `env:"SESSION_GRPC_ADDR" envDefault:"127.0.0.1:50052"`
		UserAddr    string `env:"USER_GRPC_ADDR" envDefault:"127.0.0.1:50053"`
		OrderAddr   string `env:"ORDER_GRPC_ADDR" envDefault:"127.0.0.1:50054"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
