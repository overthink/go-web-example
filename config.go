package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type HttpServerConfig struct {
	Port                int    `mapstructure:"port"`
	ListenAddress       string `mapstructure:"listen_address"`
	ReadTimeoutSeconds  int    `mapstructure:"read_timeout_s"`
	WriteTimeoutSeconds int    `mapstructure:"write_timeout_s"`
}

type PostgresConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	DbName     string `mapstructure:"dbname"`
	UserName   string `mapstructure:"username"`
	Password   string `mapstructure:"password"` // TODO: handle secrets properly
	SSLEnabled bool   `mapstructure:"ssl_enabled"`
}

type Config struct {
	HttpServer HttpServerConfig `mapstructure:"http_server"`
	Postgres   PostgresConfig   `mapstructure:"postgres"`
}

func LoadConfig() (Config, error) {
	var config Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("tasks")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return config, nil
}
