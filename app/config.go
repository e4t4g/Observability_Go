package main

import (
	"bytes"
	"github.com/spf13/viper"
	"os"
)

type Config struct {
	BotToken string `mapstructure:"bot_token"`
	BotApi   string `mapstructure:"bot_api"`
}

func InitConfig(path string) (*Config, error) {
	var cfg Config

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(file)); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil

}
