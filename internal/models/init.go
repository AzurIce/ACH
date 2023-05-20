package models

import (
	"ach/internal/utils"
	"errors"
	"fmt"
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
	_, err = GetUserByID(1)
	if err == gorm.ErrRecordNotFound {
        err := AddDefaultUser()
        if err != nil {
            log.Panicln(err)
        }
	}
}

func AddDefaultUser() error {
    defaultUser := &User{}

    defaultUser.Username = "Admin"
    defaultUser.Password = utils.EncodePassword(utils.RandStringRunes(8), utils.RandStringRunes(16))
    defaultUser.PlayerUUID = "Admin"
    defaultUser.PlayerName = "Admin"
    defaultUser.IsAdmin = true

    if err := DB.Create(&defaultUser).Error; err != nil {
        return errors.New(fmt.Sprintf("创建初始管理员账户失败: %s\n", err))
    }

    log.Println("初始管理员账户创建完成")
    log.Printf("用户名: %s\n", "Admin")
    log.Printf("密码: %s\n", defaultUser.Password)
    return nil
}
