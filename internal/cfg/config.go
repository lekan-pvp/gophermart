package cfg

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
)

type Config struct {
	RunAddress  string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI string `env:"DATABASE_URI" envDefault:"postgresql://postgres:871023@localhost:5432/gophermart_db?sslmode=disable"`
}

var instance *Config

func init() {
	log.Println("init cfg...")
	instance = &Config{}
	if err := env.Parse(instance); err != nil {
		log.Println("?")
		log.Fatal(err)
	}

	runAddress := flag.String("a", instance.RunAddress, "адрес и порт запуска сервиса")
	databaseURI := flag.String("d", instance.DatabaseURI, "URI подключения к БД")

	flag.Parse()

	instance.RunAddress = *runAddress
	instance.DatabaseURI = *databaseURI
}

func GetConfig() Config {
	return *instance
}
