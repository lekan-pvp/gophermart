package cfg

import (
	"flag"
	"github.com/caarlos0/env"
	"github.com/lekan/gophermart/internal/logger"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgresql://postgres:871023@localhost:5432/gophermart_db?sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

var instance *Config

func init() {
	log := logger.GetLogger()
	log.Info().Msg("set up config...")
	instance = &Config{}
	if err := env.Parse(instance); err != nil {
		log.Fatal().Err(err).Msg("can not parse instance")
	}

	runAddress := flag.String("a", instance.RunAddress, "адрес и порт запуска сервиса")
	databaseURI := flag.String("d", instance.DatabaseURI, "URI подключения к БД")
	accrualSystemAddress := flag.String("r", instance.AccrualSystemAddress, "адрес системы расчета начислений")

	flag.Parse()

	instance.RunAddress = *runAddress
	instance.DatabaseURI = *databaseURI
	instance.AccrualSystemAddress = *accrualSystemAddress
}

func GetConfig() Config {
	return *instance
}

func GetAccuralSystemAddress() string {
	return instance.AccrualSystemAddress
}
