package models

import (
	"ach/pkg/utils"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	db, err := gorm.Open(sqlite.Open("ach.db"), &gorm.Config{})

	if err != nil {
		log.Panicf("无法连接数据库，%s", err)
	}

	DB = db
	DB.AutoMigrate(&User{})

	// 创建初始管理员账户
	addDefaultUser()
}

func addDefaultUser() {
	_, err := GetUserByID(1)
	password := utils.RandStringRunes(8)

	if err == gorm.ErrRecordNotFound {
		defaultUser := &User{}

		defaultUser.Name = "Admin"
		defaultUser.Password = utils.EncodePassword(password, utils.RandStringRunes(16))
		defaultUser.PlayerUUID = "Admin"
		defaultUser.PlayerName = "Admin"
		defaultUser.IsAdmin = true

		if err := DB.Create(&defaultUser).Error; err != nil {
			log.Panicf("创建初始管理员账户失败: %s\n", err)
		}

		log.Println("初始管理员账户创建完成")
		log.Printf("用户名: %s\n", "Admin")
		log.Printf("密码: %s\n", password)

	}
}
