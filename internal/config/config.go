package config

import (
	"ach/internal/utils"
	// "fmt"
	"io/ioutil"
	"log"
	// "net/http"
	// "os"
	// "os/exec"
	// "path"

	"gopkg.in/yaml.v3"
)

var CONFIG_FILE_PATH = "./config.yml"

// ServerConfig ...
// If version is empty
//   - If the server is not installed then
//     install the latest version and update this
//     field.
//   - If the server is installed then
//     update this field to the current version.
// If version is not empty
//   - Install the corresponding version no matter
//     the server is already installed or not.

type ServerConfig struct {
	Dir          string `yaml:"dir"`
	// LauncherType string `yaml:"launcherType"`
	JVMOptions   string `yaml:"jvm_options"`
	Version      string `yaml:"version"`
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
			"serverName": {
				Dir:          "path/to/your/server/s/folder",
				// LauncherType: "quilt",
				JVMOptions:   "-Xms4G -Xmx4G",
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

	// 可以读取config.yml，清空 config
	log.Println("[config/ReadConfig]: Parsing...")
	err = yaml.Unmarshal(configStr, config)
	if err != nil {
		log.Println(err)
	}
	log.Print("[config/ReadConfig]: config:", config, '\n')
	return config, nil
}

func Save(config *ACHConfig) {
	log.Println("[config/SaveConfig]: Saving config to " + CONFIG_FILE_PATH + "...")
	configStr, _ := yaml.Marshal(config)
	ioutil.WriteFile(CONFIG_FILE_PATH, configStr, 0666)
}