package config

import (
	"github.com/spf13/viper"
)

const (
	ConfigFilePath = "./"
	ConfigFileName = "config"
)

type Config struct {
	HostAddr string        `mapstructure:"host_addr"`
	StartURL string        `mapstructure:"start_url"`
	Elastic  ElasticConfig `mapstructure:"elastic"`
	Redis    RedisConfig   `mapstructure:"redis"`
}

type ElasticConfig struct {
	Addr  string `mapstructure:"addr"`
	Index string `mapstructure:"index"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	QueueKey string `mapstructure:"queue_key"`
	SetKey   string `mapstructure:"set_key"`
}

func New() (Config, error) {
	vpr := viper.New()
	vpr.AddConfigPath(ConfigFilePath)
	vpr.SetConfigName(ConfigFileName)

	if err := vpr.ReadInConfig(); err != nil {
		return Config{}, err
	}

	var result Config
	if err := vpr.Unmarshal(&result); err != nil {
		return Config{}, err
	}

	return result, nil
}
