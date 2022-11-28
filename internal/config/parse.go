package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

func (c *Config) parseFromOsArgs() {

	flag.StringVar(&c.servAPIAddr, "a", c.servAPIAddr, "api server run address")
	flag.StringVar(&c.pgConnString, "d", c.pgConnString, "database connection string")
	flag.StringVar(&c.accrualAPIAddr, "r", c.accrualAPIAddr, "api accrual run address")
	flag.DurationVar(&c.orderStatusUpdateInterval, "u", c.orderStatusUpdateInterval, "order status update interval")
	flag.StringVar(&c.logLevel, "l", c.logLevel, "log level")

	flag.Parse()
}

func (c *Config) parseFromEnv() (err error) {

	envConfig := struct {
		ServAPIAddr               string        `env:"RUN_ADDRESS" toml:"RUN_ADDRESS"`
		PgConnString              string        `env:"DATABASE_URI" toml:"DATABASE_URI"`
		AccrualAPIAddr            string        `env:"ACCRUAL_SYSTEM_ADDRESS" toml:"ACCRUAL_SYSTEM_ADDRESS"`
		OrderStatusUpdateInterval time.Duration `env:"ORDER_STATUS_UPDATE_INTERVAL" toml:"ORDER_STATUS_UPDATE_INTERVAL"`
		LogLevel                  string        `env:"LOG_LEVEL" toml:"LOG_LEVEL"`
	}{}

	if err = env.Parse(&envConfig); err != nil {
		return fmt.Errorf(`parsing config from env: %w`, err)
	}

	if envConfig.ServAPIAddr != "" {
		c.servAPIAddr = envConfig.ServAPIAddr
	}

	if envConfig.PgConnString != "" {
		c.pgConnString = envConfig.PgConnString
	}

	if envConfig.AccrualAPIAddr != "" {
		c.accrualAPIAddr = envConfig.AccrualAPIAddr
	}

	if envConfig.OrderStatusUpdateInterval != 0 {
		c.orderStatusUpdateInterval = envConfig.OrderStatusUpdateInterval
	}

	if envConfig.LogLevel != "" {
		c.logLevel = envConfig.LogLevel
	}

	return nil
}
