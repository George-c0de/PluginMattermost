// Package config конфиг и работа с конфигом
package config

import "github.com/ilyakaznacheev/cleanenv"

// Config Конфигурация приложения
type Config struct {
	TarantoolHost string `env:"TARANTOOL_HOST" env-default:"localhost"`
	TarantoolPort int    `env:"TARANTOOL_PORT" env-default:"3301"`
	PortHTTP      int    `env:"PORT_HTTP" env-default:"8000"`
	Env           string `env:"ENV" env-default:"local"`
}

// MustGetConfig Загрузка конфига
func MustGetConfig() *Config {
	config := Config{}
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		panic(err)
	}
	return &config
}
