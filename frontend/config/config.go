package config

import "time"

type AppConfig struct {
	ServerPort  int           `env:"SERVER_PORT"         envDefault:"8000"`
	ReadTimeout time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"5s"`
	LogLevel    string        `env:"LOG_LEVEL"           envDefault:"info"`
}

type HTTPConfig struct {
	Timeout  time.Duration `env:"HTTP_CLIENT_TIMEOUT"   envDefault:"30s"`
	RetryMax int           `env:"HTTP_CLIENT_RETRY_MAX" envDefault:"3"`
}

type HTMLConfig struct {
	Css string `env:"CSS" envDefault:"static/styles/index.css"`
}
