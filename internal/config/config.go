package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	TarantoolHost string `env:"TARANTOOL_HOST" env-default:"localhost"`
	TarantoolPort int    `env:"TARANTOOL_PORT" env-default:"3301"`
	PortHttp      int    `env:"PORT_HTTP" env-default:"8000"`
	Env           string `env:"ENV" env-default:"local"`
}

func MustGetConfig() *Config {
	config := Config{}
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		panic(err)
	}
	return &config
}
