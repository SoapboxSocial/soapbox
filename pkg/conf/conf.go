// Package conf contains utility functions for loading and parsing configuration files.
package conf

import (
	"os"

	"github.com/spf13/viper"
)

// PostgresConf describes a default configuration for the postgres database.
type PostgresConf struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSL      string `mapstructure:"ssl"`
}

// PostgresConf describes a default configuration for the redis.
type RedisConf struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// Load opens and parses a configuration file.
func Load(file string, conf interface{}) error {
	_, err := os.Stat(file)
	if err != nil {
		return err
	}

	viper.SetConfigFile(file)
	viper.SetConfigType("toml")

	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.GetViper().Unmarshal(conf)
	if err != nil {
		return err
	}

	return nil
}
