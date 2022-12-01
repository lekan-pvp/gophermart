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

func New() *Config {
	var runAddress string
	var databaseURI string
	var accrualSystemAddress string

	var once sync.Once
	once.Do(func() {
		log := logger.New()
		singleton = &Config{}
		if err := env.Parse(singleton); err != nil {
			log.Fatal().Err(err).Msg("can not parse Config")
		}

		flag.StringVar(&runAddress, "a", singleton.RunAddress, "адрес и порт запуска сервиса")
		flag.StringVar(&databaseURI, "d", singleton.DatabaseURI, "URI подключения к БД")
		flag.StringVar(&accrualSystemAddress, "r", singleton.AccrualSystemAddress, "адрес системы расчета начислений")

		flag.Parse()
		singleton.RunAddress = runAddress
		singleton.DatabaseURI = databaseURI
		singleton.AccrualSystemAddress = accrualSystemAddress
	})
	return singleton
}

func GetAccrualSystemAddress() string {
	return singleton.AccrualSystemAddress
}
