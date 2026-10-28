package config

import "github.com/caarlos0/env/v10"

type Config struct {
	GRPC struct {
		Port string `env:"GRPC_PORT" envDefault:"50052"`
	}
	Repository struct {
		Driver string `env:"SESSION_REPOSITORY" envDefault:"memory"`
	}
	Postgres struct {
		DSN string `env:"SESSION_POSTGRES_DSN" envDefault:"postgres://session:session@127.0.0.1:5432/session?sslmode=disable"`
	}
	Context struct {
		PreviousChatsLimit int `env:"SESSION_PREVIOUS_CHATS_LIMIT" envDefault:"4"`
	}
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
