package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTPPort string `required:"true" envconfig:"HTTP_PORT"`
	DatabaseConfig
	KafkaConfig
}

func NewConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

type DatabaseConfig struct {
	User     string `required:"true" envconfig:"MYSQL_USER"`
	Password string `required:"true" envconfig:"MYSQL_PASSWORD"`
	Host     string `required:"true" envconfig:"MYSQL_HOST"`
	Port     string `required:"true" envconfig:"MYSQL_PORT"`
	Name     string `required:"true" envconfig:"MYSQL_DATABASE"`
}

type KafkaConfig struct {
	Brokers []string `required:"true" envconfig:"KAFKA_BROKERS"`
}

type TestDatabaseConfig struct {
	User     string `required:"true" envconfig:"MYSQL_USER"`
	Password string `required:"true" envconfig:"MYSQL_PASSWORD"`
	Host     string `required:"true" envconfig:"MYSQL_HOST"`
	Port     string `required:"true" envconfig:"MYSQL_PORT"`
	Name     string `required:"true" envconfig:"MYSQL_TEST_DATABASE"`
}

func NewTestDatabaseConfig() (*TestDatabaseConfig, error) {
	var cfg TestDatabaseConfig
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
