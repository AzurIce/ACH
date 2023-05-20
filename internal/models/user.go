package models

import (
	"ach/internal/utils"
	"log"
	"strings"
)

type User struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	PlayerUUID string `json:"uuid" gorm:"unique"`     // 玩家 UUID
	PlayerName string `json:"player_name"`            // 玩家名(用户名)
	Username   string `json:"username" gorm:"unique"` // 用户名
	Password   string `json:"-"`                      // 密码
	IsAdmin    bool   `json:"is_admin"`               // 是否是管理员
}

func (user *User) CheckPassword(password string) bool {
	salt := strings.Split(user.Password, ":")[0]
	log.Printf("用户密码为: %s", user.Password)
	log.Printf("盐: %s", salt)

	return user.Password == utils.EncodePassword(password, salt)
}

func CreateUser(uuid string, name string) (uint, error) {
	log.Printf("正在创建<User>(PlayerUUID = %s, PlayerName = %s)...", uuid, name)
	user := User{PlayerUUID: uuid, PlayerName: name}

	res := DB.Create(&user)
	if res.Error == nil {
		log.Printf("查找完成: <User>(Username = %s, PlayerUUID = %s, PlayerName = %s)", user.Username, user.PlayerUUID, user.PlayerName)
	}
	return user.ID, res.Error
}

func GetUserByUUID(uuid string) (User, error) {
	log.Printf("正在查找<User>(PlayerUUID = %s)...", uuid)
	var user User

	res := DB.Where("player_uuid = ?", uuid).First(&user)
	if res.Error != nil {
		log.Printf("查找失败: %s", res.Error)
		return user, res.Error
	}
	log.Printf("查找完成: <User>(Username = %s, PlayerUUID = %s, PlayerName = %s)", user.Username, user.PlayerUUID, user.PlayerName)
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
	log.Printf("查找完成: <User>(Username = %s, PlayerUUID = %s, PlayerName = %s)", user.Username, user.PlayerUUID, user.PlayerName)
	return user, nil
}

func DeleteUserById(id uint) error {
	log.Printf("正在删除<User>(ID = %d)...", id)

    res := DB.Delete(&User{}, id)
    if res.Error != nil {
        log.Printf("删除失败: %s", res.Error)
        return res.Error
    }
    
    log.Printf("删除完成: <User>(id = %v)", id)
    return nil
}

func GetUserByUsername(username string) (User, error) {
	log.Printf("正在查找<User>(Username = %s)...", username)
	var user User

	res := DB.Where("username = ?", username).First(&user)
	if res.Error != nil {
		log.Printf("查找失败: %s", res.Error)
		return user, res.Error
	}
	log.Printf("查找完成: <User>(Username = %s, PlayerUUID = %s, PlayerName = %s)", user.Username, user.PlayerUUID, user.PlayerName)
	return user, nil
}

func GetUserList() ([]User, error) {
	log.Println("正在获取所有 User...")
	var userList = make([]User, 0)

	res := DB.Find(&userList)
	if res.Error != nil {
		log.Printf("获取失败: %s", res.Error)
		return userList, res.Error
	}
	log.Printf("获取完成：共 %d 条数据", len(userList))

	return userList, nil
}
