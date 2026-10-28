package config

import "github.com/caarlos0/env/v10"

type Config struct {
	Mode         string `env:"PROVIDER_MODE" envDefault:"demo"`
	RegistryFile string `env:"PROVIDER_CONFIG_FILE" envDefault:"configs/providers.json"`
	GRPC         struct {
		Host string `env:"GRPC_HOST" envDefault:"127.0.0.1"`
		Port string `env:"GRPC_PORT" envDefault:"50058"`
	}
	RegistryPath string `env:"PROVIDERS_CONFIG"`
	POIPath      string `env:"POI_INDEX_PATH" envDefault:"configs/poi/kazan.json"`
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
