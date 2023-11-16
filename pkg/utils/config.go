package utils

import (
	"github.com/spf13/viper"
)

// LoadConfig charge la config toml
func LoadConfig(confDirectory, confFile, migrationDir string) {
	viper.SetDefault("migrationsDir", migrationDir)
	viper.SetConfigName(confFile)
	viper.SetConfigType("toml")
	viper.AddConfigPath(confDirectory)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
