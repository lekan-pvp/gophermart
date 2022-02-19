package config

import (
	"flag"
	"github.com/caarlos0/env"
	"github.com/lekan/gophermart/internal/logger"
	"sync"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:":8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgresql://postgres:871023@localhost:5432/gophermart_db?sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8000"`
}

var singleton *Config

var instance *Config

func New() *Config {
	var once sync.Once
	once.Do(func() {
		log := logger.New()
		singleton = &Config{}
		if err := env.Parse(singleton); err != nil {
			log.Fatal().Err(err).Msg("can not parse Config")
		}
		runAddress := flag.String("a", instance.RunAddress, "адрес и порт запуска сервиса")
		databaseURI := flag.String("d", instance.DatabaseURI, "URI подключения к БД")
		accrualSystemAddress := flag.String("r", instance.AccrualSystemAddress, "адрес системы расчета начислений")

		flag.Parse()
		singleton.RunAddress = *runAddress
		singleton.DatabaseURI = *databaseURI
		singleton.AccrualSystemAddress = *accrualSystemAddress
	})
	return singleton
}

func GetAccrualSystemAddress() string {
	return singleton.AccrualSystemAddress
}
