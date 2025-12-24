package config

import (
	"github.com/spf13/viper"
)

func InitConfigWithViper(filePath string) {
	err := loadConfig(filePath)
	if err != nil {
		panic(err)
	}
}

func loadConfig(filePath string) error {
	viper.SetConfigFile(filePath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}
