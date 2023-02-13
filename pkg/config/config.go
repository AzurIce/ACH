package config

import (
	"ach/pkg/utils"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var CONFIG_FILE_PATH = "./config.yml"

// ServerConfig ...
type ServerConfig struct {
	ExecOptions string `yaml:"execOptions"`
	ExecPath    string `yaml:"execPath"`
}

// ACHConfig ...
type ACHConfig struct {
	CommandPrefix    string                  `yaml:"command_prefix"`
	BackupDir        string                  `yaml:"backup_dir"`
	Servers          map[string]ServerConfig `yaml:"servers"`
	JWTSigningString string                  `yaml:"jwt_signing_string"`
}

func DefaultACHConfig() *ACHConfig {
	return &ACHConfig{
		CommandPrefix:    "#",
		BackupDir:        "./Backups",
		JWTSigningString: utils.RandStr(6),
		Servers: map[string]ServerConfig{
			"serverName1": {
				ExecOptions: "-Xms4G -Xmx4G",
				ExecPath:    "path/to/your/server/s/exec/jar/file",
			},
		},
	}
}

func ReadConfig() (*ACHConfig, error) {
	log.Println("[config/ReadConfig]: Reading " + CONFIG_FILE_PATH + "...")
	configStr, err := ioutil.ReadFile(CONFIG_FILE_PATH)
	if err != nil { // 读取文件发生错误
		return DefaultACHConfig(), err
	}

	config := &ACHConfig{}

	// 可以读取config.yml，清空ach.config
	log.Println("[config/ReadConfig]: Parsing...")
	err = yaml.Unmarshal(configStr, config)
	if err != nil {
		log.Println(err)
	}
	log.Print("[config/ReadConfig]: config:", config, '\n')
	return config, nil
}

func SaveConfig(config *ACHConfig) {
	log.Println("[config/SaveConfig]: Saving config to " + CONFIG_FILE_PATH + "...")
	configStr, _ := yaml.Marshal(config)
	ioutil.WriteFile(CONFIG_FILE_PATH, configStr, 0666)
}
