package bootstrap

import (
	"ach/config"
	"os"
)

var Config *config.ACHConfig

func InitConfig() {
	Config = config.DefaultACHConfig()

	err := Config.Read()
	if err != nil {
		if os.IsNotExist(err) { // 文件不存在，创建并写入默认配置
			println("[ACH]: Cannot find config.yml, creating...")
			Config.Save()
			println("[ACH]: Successful created config.yml, please complete the config.")
		}
		os.Exit(1)
	}
}
