package config

import (
	"time"

	"github.com/rs/zerolog/log"
)

type Config struct {
	servAddr                  string
	pgConnString              string
	accrualAddr               string
	orderStatusUpdateInterval time.Duration
}

func New(options ...string) (*Config, error) {
	log.Debug().Strs("options", options).Msg("config.New started")
	var err error
	cfg := &Config{}
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("config.New ended")
		} else {
			log.Debug().Str("config", cfg.String()).Msg("config.New ended")
		}
	}()

	cfg.orderStatusUpdateInterval = time.Second * 1

	for _, opt := range options {
		switch opt {
		case WithFlag:
			cfg.parseFromOsArgs()
		case WithEnv:
			if err = cfg.parseFromEnv(); err != nil {
				return nil, err
			}
		}
	}

	return cfg, nil
}

func (c *Config) ServAddr() string {
	return c.servAddr
}

func (c *Config) PgConnString() string {
	return c.pgConnString
}

func (c *Config) AccrualAddr() string {
	return c.accrualAddr
}

func (c *Config) OrderStatusUpdateInterval() time.Duration {
	return c.orderStatusUpdateInterval
}

func (c *Config) String() string {
	if c == nil {
		return "config is nil pointer"
	}
	return "servAddr: " + c.servAddr +
		" accrualAddr: " + c.accrualAddr +
		" orderStatusUpdateInterval" + c.orderStatusUpdateInterval.String()
}
