package models

import (
	"ach/pkg/utils"
	"log"
	"strings"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	PlayerUUID string // 玩家 UUID
	PlayerName string // 玩家名(用户名)
	Name       string // 用户名
	Password   string // 密码
	IsAdmin    bool   // 是否是管理员
}

func (user *User) CheckPassword(password string) bool {
	salt := strings.Split(user.Password, ":")[0]
	log.Printf("用户密码为: %s", user.Password)
	log.Printf("盐: %s", salt)

	return user.Password == utils.EncodePassword(password, salt)
}

func CreateUser(uuid string, name string) uint {
	log.Printf("正在创建<User>(PlayerUUID = %s, PlayerName = %s)...", uuid, name)
	user := User{PlayerUUID: uuid, PlayerName: name}

	res := DB.Create(&user)
	if res.Error == nil {
		log.Printf("查找完成: <User>(Name = %s, PlayerUUID = %s, PlayerName = %s)", user.Name, user.PlayerUUID, user.PlayerName)
	}
	return user.ID
}

func GetUserByUUID(uuid string) (User, error) {
	log.Printf("正在查找<User>(PlayerUUID = %s)...", uuid)
	var user User

	res := DB.Where("player_uuid = ?", uuid).First(&user)
	if res.Error != nil {
		log.Printf("查找失败: %s", res.Error)
		return user, res.Error
	}
	log.Printf("查找完成: <User>(Name = %s, PlayerUUID = %s, PlayerName = %s)", user.Name, user.PlayerUUID, user.PlayerName)
	return user, nil
}

func GetUserByID(id uint) (User, error) {
	log.Printf("正在查找<User>(ID = %d)...", id)
	var user User

	res := DB.First(&user, id)
	if res.Error != nil {
		log.Printf("查找失败: %s", res.Error)
		return user, res.Error
	}
	log.Printf("查找完成: <User>(Name = %s, PlayerUUID = %s, PlayerName = %s)", user.Name, user.PlayerUUID, user.PlayerName)
	return user, nil
}

func GetUserByName(name string) (User, error) {
	log.Printf("正在查找<User>(Name = %s)...", name)
	var user User

	res := DB.Where("name = ?", name).First(&user)
	if res.Error != nil {
		log.Printf("查找失败: %s", res.Error)
		return user, res.Error
	}
	log.Printf("查找完成: <User>(Name = %s, PlayerUUID = %s, PlayerName = %s)", user.Name, user.PlayerUUID, user.PlayerName)
	return user, nil
}

func GetUserList() ([]User, error) {
	log.Println("正在获取所有 User...")
	var userList = make([]User, 0)

	res := DB.Select("player_uuid", "player_name", "name", "is_admin").Find(&userList)
	if res.Error != nil {
		log.Printf("获取失败: %s", res.Error)
		return userList, res.Error
	}
	log.Printf("获取完成：共 %d 条数据", len(userList))

	return userList, nil
}
