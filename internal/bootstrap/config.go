package bootstrap

import (
	"ach/internal/config"
	"ach/internal/utils"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
)

var Config *config.ACHConfig

func init() {
	var err error
	log.Println("[bootstrap/InitConfig]: Initializing config...")
	Config = config.DefaultACHConfig()

	Config, err = config.ReadConfig()
	if err != nil {
		if os.IsNotExist(err) { // 文件不存在，创建并写入默认配置
			println("[bootstrap/InitConfig]: Cannot find config.yml, creating default config...")
			config.Save(Config)
			println("[bootstrap/InitConfig]: Successfully created config.yml, please complete the config.")
		}
		os.Exit(1)
	}

	CheckConfig()
	
	// log.Print("[bootstrap/InitConfig]: config: ", Config, '\n')
	os.Mkdir(Config.BackupDir, 0666)
	os.Mkdir(path.Join(Config.BackupDir, "backups"), 0666)
	os.Mkdir(path.Join(Config.BackupDir, "snapshots"), 0666)
}

func CheckConfig() {
	for serverName, serverConfig := range Config.Servers {
		// if len(serverConfig.LauncherType) == 0 {
		// 	log.Panicln("LauncherType cannot be nil")
		// }
		// if serverConfig.LauncherType != "vanilla" &&
		// 	serverConfig.LauncherType != "fabric" &&
		// 	serverConfig.LauncherType != "quilt" {
		// 	log.Panicln("LauncherType can only be one of vanilla, fabric and quilt")
		// }
		if _, err := os.Stat(serverConfig.Dir); err != nil {
			log.Panicln(err)
		} else {
			needInstall := false
			downloadServer := false
			if _, err := os.Stat(path.Join(serverConfig.Dir, "server.jar")); err != nil {
				log.Printf("[%v]: server.jar not found.", serverName)
				needInstall = true
				downloadServer = true
			}
			if _, err := os.Stat(path.Join(serverConfig.Dir, "quilt-server-launch.jar")); err != nil {
				log.Printf("[%v]: quilt-server-launch.jar not found.", serverName)
				needInstall = true
			}
			if _, err := os.Stat(path.Join(serverConfig.Dir, "versions", serverConfig.Version)); err != nil {
				log.Printf("[%v]: version cannot match.", serverName)
				needInstall = true
				downloadServer = true
			}
			downloadServer = true
			
			if needInstall {
				utils.DeletePath(path.Join(serverConfig.Dir, ".quilt"))
				utils.DeletePath(path.Join(serverConfig.Dir, "config"))
				utils.DeletePath(path.Join(serverConfig.Dir, "libraries"))
				utils.DeletePath(path.Join(serverConfig.Dir, "versions"))
				utils.DeletePath(path.Join(serverConfig.Dir, "quilt-server-launch.jar"))
				utils.DeletePath(path.Join(serverConfig.Dir, "server.jar"))
				if _, err := os.Stat("./quilt-installer-0.5.0.jar"); err != nil {
					log.Printf("quilt-installer-0.5.0.jar not found, downloading...")
					DownloadQuiltInstaller()
				}

				log.Printf("Installing server using quilt-installer-0.5.0.jar...")
				// java -jar quilt-installer-0.5.0.jar install server 1.19.4 --download-server
				args := []string{"-jar", "quilt-installer-0.5.0.jar", "install", "server"}
				if len(serverConfig.Version) != 0 {
					args = append(args, serverConfig.Version)
				} else {
					serverConfig.Version = "1.19.4"
					args = append(args, "1.19.4")
					Config.Servers[serverName] = serverConfig
					config.Save(Config)
				}
				if downloadServer {
					args = append(args, "--download-server")
				}
				args = append(args, fmt.Sprintf("--install-dir=%v", serverConfig.Dir))
				output, err := exec.Command("java", args...).Output()
				if err != nil {
					log.Println(err)
					log.Panicf("Install failed: %v\n%v\n", err, string(output))
				} else {
					log.Println(string(output))
				}
			}
		}
	}
}

func DownloadQuiltInstaller() {
	url := "https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-installer/0.5.0/quilt-installer-0.5.0.jar"
	res, err := http.Get(url)
    if err != nil {
        log.Panicf("http.Get -> %v\n", err)
    }
    data, err := ioutil.ReadAll(res.Body)
    if err != nil {
        log.Panicf("ioutil.ReadAll -> %s\n", err.Error())
    }
    defer res.Body.Close()
    if err = ioutil.WriteFile("./quilt-installer-0.5.0.jar", data, 0777); err != nil {
        log.Panicf("Error Saving: %v\n", err)
    }
}