package config

import "github.com/spf13/viper"

type Config struct {
	Server struct {
		Port       string `mapstructure:"port"`
		JWTSecret  string `mapstructure:"jwt_secret"`
		CORSOrigin string `mapstructure:"cors_origin"`
	} `mapstructure:"server"`
	Database struct {
		DSN          string `mapstructure:"dsn"`
		MaxOpenConns int    `mapstructure:"max_open_conns"`
		MaxIdleConns int    `mapstructure:"max_idle_conns"`
	} `mapstructure:"database"`
}

func Load(path string) (Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if path == "" {
		viper.SetConfigName("config")
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
