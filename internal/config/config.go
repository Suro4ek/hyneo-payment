package config

import (
	"hyneo-payment/pkg/logging"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config is application config
type Config struct {
	MySQL  MySQL  `yaml:"mysql"`
	SECRET string `yaml:"secret_key" env:"SECRET_KEY"`
	IP     string `yaml:"ip" env:"IP"`
}

type MySQL struct {
	Host     string `yaml:"host" env:"MYSQL_HOST"`
	Port     string `yaml:"port" env:"MYSQL_PORT"`
	User     string `yaml:"user" env:"MYSQL_USER"`
	Password string `yaml:"pass" env:"MYSQL_PASS"`
	DB       string `yaml:"db" env:"MYSQL_DB"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("read application config")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})
	return instance
}
