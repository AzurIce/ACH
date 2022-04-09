package models

import (
	"log"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UUID string // 玩家 UUID
	Name string // 玩家名
}

func CreateUser(uuid string, name string) uint {
	log.Printf("正在创建<User>(UUID = %s, name = %s)...", uuid, name)
	user := User{UUID: uuid, Name: name}

	res := DB.Create(&user)
	if (res.Error == nil) {
		log.Printf("创建完成: <User>(UUID = %s, Name = %s)", user.UUID, user.Name)
	}
	return user.ID
}

func GetUserByUUID(uuid string) (User, error) {
	defer func() {
		rec := recover()
		log.Println(rec)
	}()
	log.Printf("正在查找<User>(UUID = %s)...", uuid)
	var user User

	res := DB.Where("uuid = ?", uuid).First(&user)
	if res.Error != nil {
		log.Printf("查找失败: %s", res.Error)
		return user, res.Error
	}
	log.Printf("查找完成: <User>(UUID = %s, Name = %s)", user.UUID, user.Name)
	return user, nil
}
