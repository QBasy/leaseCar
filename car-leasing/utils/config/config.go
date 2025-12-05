package config

import (
	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MeiliSearchConfig struct {
	URL    string `mapstructure:"url"`
	APIKey string `mapstructure:"api_key"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type Config struct {
	Server      ServerConfig       `mapstructure:"server"`
	Database    DatabaseConfig     `mapstructure:"database"`
	Redis       RedisConfig        `mapstructure:"redis"`
	MeiliSearch MeiliSearchConfig  `mapstructure:"meilisearch"`
}

// Load loads configuration from file
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Override with environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadKey loads a specific key from config
func LoadKey(configPath, key string, target interface{}) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.UnmarshalKey(key, target)
}
