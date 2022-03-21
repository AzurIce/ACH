package core

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
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

func (ach *ACHCore) readConfig() error {
	configYaml, err := ioutil.ReadFile("./config.yml")
	if err != nil { // 读取文件发生错误
		return err
	}
	// 可以读取config.yml，清空ach.config
	ach.config = &ACHConfig{}
	err = yaml.Unmarshal(configYaml, ach.config)
	// fmt.Println(ach.config)
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (ach *ACHCore) saveConfig() {
	data, _ := yaml.Marshal(ach.config)
	ioutil.WriteFile("./config.yml", data, 0666)
}
