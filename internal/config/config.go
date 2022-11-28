package config

import (
	"time"
)

type Config struct {
	servAPIAddr  string
	pgConnString string

	accrualAPIAddr            string
	accrualGetOrder           string
	orderStatusUpdateInterval time.Duration

	logLevel string
}

func New(options ...string) (newCfg *Config, err error) {

	newCfg = &Config{}

	for _, opt := range options {
		switch opt {
		case WithFlag:
			newCfg.parseFromOsArgs()
		case WithEnv:
			if err = newCfg.parseFromEnv(); err != nil {
				return nil, err
			}
		}
	}

	newCfg.setDefaultIfNotConfigured()

	newCfg.accrualGetOrder = newCfg.accrualAPIAddr + "/api/orders/{number}"

	return newCfg, nil
}

func (c *Config) setDefaultIfNotConfigured() {

	if c.servAPIAddr == "" {
		c.servAPIAddr = "localhost:8081"
	}

	if c.accrualAPIAddr == "" {
		c.accrualAPIAddr = "localhost:8080"
	}

	if c.orderStatusUpdateInterval == 0 {
		c.orderStatusUpdateInterval = time.Second * 5
	}

	if c.logLevel == "" {
		c.logLevel = "info"
	}

}

func (c *Config) ServAPIAddr() string {
	return c.servAPIAddr
}

func (c *Config) PgConnString() string {
	return c.pgConnString
}

func (c *Config) AccrualAPIAddr() string {
	return c.accrualAPIAddr
}

func (c *Config) AccrualGetOrder() string {
	return c.accrualGetOrder
}

func (c *Config) OrderStatusUpdateInterval() time.Duration {
	return c.orderStatusUpdateInterval
}

func (c *Config) LogLevel() string {
	return c.logLevel
}

func (c *Config) String() string {
	if c == nil {
		return "config is nil pointer"
	}
	return "servAPIAddr: " + c.servAPIAddr +
		" accrualAPIAddr: " + c.accrualAPIAddr +
		" accrualGetOrder: " + c.accrualGetOrder +
		" orderStatusUpdateInterval" + c.orderStatusUpdateInterval.String() +
		" logLevel" + c.LogLevel()
}
