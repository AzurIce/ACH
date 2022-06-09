package config

import (
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
	CommandPrefix string                  `yaml:"command_prefix"`
	BackupDir     string                  `yaml:"backup_dir"`
	Servers       map[string]ServerConfig `yaml:"servers"`
}

func DefaultACHConfig() *ACHConfig {
	return &ACHConfig{
		CommandPrefix: "#",
		BackupDir:     "./Backups",
		Servers: map[string]ServerConfig{
			"serverName1": {
				ExecOptions: "-Xms4G -Xmx4G",
				ExecPath:    "path/to/your/server/s/exec/jar/file",
			},
		},
	}
}

func ReadConfig() (*ACHConfig, error) {
	config := &ACHConfig{}

	log.Println("[config/ReadConfig]: Reading " + CONFIG_FILE_PATH + "...")
	configYaml, err := ioutil.ReadFile(CONFIG_FILE_PATH)
	if err != nil { // 读取文件发生错误
		return DefaultACHConfig(), err
	}
	
	// 可以读取config.yml，清空ach.config
	log.Println("[config/ReadConfig]: Parsing...")
	err = yaml.Unmarshal(configYaml, config)
	if err != nil {
		log.Println(err)
	}
	log.Print("[config/ReadConfig]: config:", config, '\n')
	return config, nil
}

func SaveConfig(config *ACHConfig)  {
	log.Println("[config/SaveConfig]: Saving config to " + CONFIG_FILE_PATH + "...")
	data, _ := yaml.Marshal(config)
	ioutil.WriteFile(CONFIG_FILE_PATH, data, 0666)
}
