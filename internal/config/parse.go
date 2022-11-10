package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

func (c *Config) parseFromOsArgs() {
	log.Debug().Msg("config.parseFromOsArgs started")
	defer log.Debug().Msg("config.parseFromOsArgs ended")

	flag.StringVar(&c.servAddr, "a", c.servAddr, "service start address and port")
	flag.StringVar(&c.pgConnString, "d", c.pgConnString, "database connection address")
	flag.StringVar(&c.accrualAddr, "r", c.accrualAddr, "address of the accrual system")
	flag.DurationVar(&c.orderStatusUpdateInterval, "u", c.orderStatusUpdateInterval, "order status update interval")

	flag.Parse()
}

func (c *Config) parseFromEnv() error {
	log.Debug().Msg("config.parseFromEnv started")
	var err error
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("config.parseFromEnv ended")
		} else {
			log.Debug().Msg("config.parseFromEnv ended")
		}
	}()

	envConfig := struct {
		ServAddr                  string        `env:"RUN_ADDRESS" toml:"RUN_ADDRESS"`
		PgConnString              string        `env:"DATABASE_URI" toml:"DATABASE_URI"`
		AccrualAddr               string        `env:"ACCRUAL_SYSTEM_ADDRESS" toml:"ACCRUAL_SYSTEM_ADDRESS"`
		OrderStatusUpdateInterval time.Duration `env:"ORDER_STATUS_UPDATE_INTERVAL" toml:"ORDER_STATUS_UPDATE_INTERVAL"`
	}{}

	if err = env.Parse(&envConfig); err != nil {
		return err
	}

	c.servAddr = envConfig.ServAddr
	c.pgConnString = envConfig.PgConnString
	c.accrualAddr = envConfig.AccrualAddr
	if envConfig.OrderStatusUpdateInterval != time.Second*0 {
		c.orderStatusUpdateInterval = envConfig.OrderStatusUpdateInterval
	}

	return nil
}
