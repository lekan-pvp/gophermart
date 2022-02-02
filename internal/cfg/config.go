package cfg

import (
	"flag"
	"github.com/caarlos0/env"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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

	conn, err := gorm.Open(postgres.Open(*databaseURI))
	if err != nil {
		log.Fatal(err)
	}
	db = conn
	//db.Debug().AutoMigrate()
}

func GetDB() *gorm.DB {
	return db
}

func GetConfig() Config {
	return instance
}
