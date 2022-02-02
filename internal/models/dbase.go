package models

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lekan/gophermart/internal/cfg"
	"gorm.io/gorm"
)

func GetDB() *gorm.DB {
	return cfg.GetDB()
}
