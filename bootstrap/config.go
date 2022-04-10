package bootstrap

import (
	"ach/config"
	"log"
	"os"
)

var Config *config.ACHConfig

func InitConfig() {
	var err error
	log.Println("[Bootstrap/config]: Initializing config...")
	Config = config.DefaultACHConfig()

	Config, err = config.ReadConfig()
	if err != nil {
		if os.IsNotExist(err) { // 文件不存在，创建并写入默认配置
			println("[ACH]: Cannot find config.yml, creating...")
			Config.Save()
			println("[ACH]: Successful created config.yml, please complete the config.")
		}
		os.Exit(1)
	}
	log.Print("[Bootstrap/config]: config: ", Config, '\n')
}
