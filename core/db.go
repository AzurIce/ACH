package core

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (ach *ACHCore) initDB() {
	var err error
	ach.db, err = gorm.Open(sqlite.Open("ach.db"), &gorm.Config{})
	if err != nil {
		log.Panicln("Database init failed")
	}
}