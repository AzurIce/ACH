package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

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

	log.Println("[config]: Reading config.yml...")
	configYaml, err := ioutil.ReadFile("./config.yml")
	if err != nil { // 读取文件发生错误
		return DefaultACHConfig(), err
	}
	// log.Print("[config]: content:", string(configYaml), '\n')
	// 可以读取config.yml，清空ach.config
	log.Println("[config]: Parsing...")
	err = yaml.Unmarshal(configYaml, config)
	// fmt.Println(ach.config)
	if err != nil {
		log.Println(err)
	}
	log.Print("[config]: config:", config, '\n')
	return config, nil
}

func (config *ACHConfig) Save() {
	data, _ := yaml.Marshal(config)
	ioutil.WriteFile("./config.yml", data, 0666)
}
