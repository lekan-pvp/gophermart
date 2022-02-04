package cfg

import (
	"flag"
	"github.com/caarlos0/env"
	"github.com/lekan/gophermart/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type Config struct {
	RunAddress  string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI string `env:"DATABASE_URI" envDefault:"postgres://postgres:871023@localhost:5432/gophermart_db"`
}

var db *gorm.DB
var instance Config

func init() {
	log.Println("init cfg...")
	instance = Config{}
	if err := env.Parse(instance); err != nil {
		log.Fatal(err)
	}

	runAddress := flag.String("a", instance.RunAddress, "адрес и порт запуска сервиса")
	databaseURI := flag.String("d", instance.DatabaseURI, "URI подключения к БД")

	flag.Parse()

	instance.RunAddress = *runAddress
	instance.DatabaseURI = *databaseURI
	db, err := gorm.Open(postgres.Open(instance.DatabaseURI), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&models.Account{})
	if err != nil {
		panic("failed to migrate scheme")
	}
}

func GetDB() *gorm.DB {
	return db
}

func GetConfig() Config {
	return instance
}
