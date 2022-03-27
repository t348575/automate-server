package config

import (
	"github.com/automate/automate-server/utils-go"
	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port            string `env:"LISTEN_ADDR" envDefault:":3000"`
	Timeout         uint64 `env:"TIMEOUT" envDefault:"10"`
	ReadBufferSize  int    `env:"READ_BUFFER_SIZE" envDefault:"4096"`
	BodyLimit       int    `env:"BODY_LIMIT" envDefault:"1048576"`
	AppName         string `env:"APP_NAME" envDefault:"Automate"`
	IsProduction    bool   `env:"PRODUCTION"`
	InternalService string `env:"INTERNAL_SERVICE"`
}

func Parse() (*Config, error) {
	cfg := Config{
		IsProduction: utils.ParseFlags(),
	}

	if err := env.Parse(&cfg); err != nil {
		log.Panic().Err(err).Msg("Failed to parse env config")
	}

	return &cfg, nil
}
