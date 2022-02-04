package models

import (
	"github.com/lekan/gophermart/internal/cfg"
	"gorm.io/gorm"
)

func GetDB() *gorm.DB {
	return cfg.GetDB()
}
