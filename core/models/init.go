package models

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var (
		db *gorm.DB
		err error
	)

	db, err = gorm.Open(sqlite.Open("ach.db"), &gorm.Config{})

	if err != nil {
		log.Panicf("无法连接数据库，%s", err)
	}

	DB = db

	DB.AutoMigrate(&User{})
}