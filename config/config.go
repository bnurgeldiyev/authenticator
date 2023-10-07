package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

var Conf *Config

type (
	Config struct {
		Http
		Database
		Redis
		Jwt
	}

	Http struct {
		Host string `env-required:"true" env:"HTTP_HOST"`
		Port string `env-required:"true" env:"HTTP_PORT"`
	}

	Database struct {
		Host     string `env-required:"true" env:"DB_HOST"`
		Port     int    `env-required:"true" env:"DB_PORT"`
		User     string `env-required:"true" env:"DB_USER"`
		Password string `env-required:"true" env:"DB_PASSWORD"`
		Database string `env-required:"true" env:"DB_NAME"`
		PoolMax  int    `env-required:"true" env:"DB_POOL_MAX"`
		URL      string `env-required:"true" env:"DB_URL"`
	}

	Jwt struct {
		AccessTokenExpiry  int    `env-required:"true" env:"ACCESS_TOKEN_EXPIRY"`  // minute
		RefreshTokenExpiry int    `env-required:"true" env:"REFRESH_TOKEN_EXPIRY"` // minute
		Secret             string `env-required:"true" env:"TOKEN_SECRET"`
	}

	Redis struct {
		Host string `env-required:"true" env:"REDIS_HOST"`
		Port int    `env-required:"true" env:"REDIS_PORT"`
		URL  string `env-required:"true" env:"REDIS_CONN"`
	}
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func NewConfig() (*Config, error) {

	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	Conf = cfg

	return cfg, nil
}
