package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"ach/internal/config"
	"ach/internal/utils"
)

// Command ...
type Command struct {
	Cmd  string
	Args []string
}

// Cmds ...
var Cmds = map[string]func(*Server, ...string) error{
	"backup":  backup,
	"bksnap":  bksnap,
	"start":   start,
	"restart": restart,
}

// The length of the "args" argument is >= 1, if no args input, it will be [""]

func bksnap(server *Server, args... string) error {
	if args[0] == "" || args[0] == "list" {
		snapshotList := server.GetSnapshotList()
		for i, snapshot := range snapshotList {
			fmt.Printf("[%v] %s\n", i, snapshot.FileName)
			server.Write(fmt.Sprintf("say [%v] %s", i, snapshot.FileName))
		}
	} else if args[0] == "make" {
		comment := ""
		if len(args) > 1 {
			comment = strings.Join(args[1:], " ")
		}
		if err := server.MakeSnapshot(comment); err != nil {
			return err
		}
	} else if args[0] == "load" {
		i, err := strconv.Atoi(strings.Join(args[1:], ""))
		if err == nil {
			server.LoadSnapshot(i)
		}
	}
	return nil
}

func backup(server *Server, args... string) error {
	if args[0] == "make" {
		comment := utils.GetTimeStamp()
		if len(args) > 1 {
			comment += " " + strings.Join(args[1:], " ")
		}
		dst := path.Join(Config.BackupDir, "backups", fmt.Sprintf("%s - %s", server.Name, comment))
		if err := server.MakeBackup(dst); err != nil {
			log.Println(err)
			return err
		}
	} else if args[0] == "" || args[0] == "list" {
		backupList := server.GetBackupList(path.Join(Config.BackupDir, "backups"))
		// log.Printf("[%s/INFO]: Listing backup.\n", server.ServerName)
		for i, backup := range backupList {
			fmt.Printf("[%v] %s\n", i, backup.FileName)
			server.Write(fmt.Sprintf("say [%v] %s", i, backup.FileName))
		}
	} else if args[0] == "load" {
		i, err := strconv.Atoi(strings.Join(args[1:], ""))
		if err == nil {
			backupName := server.GetBackupList(path.Join(Config.BackupDir, "backups"))[i].FileName
			server.LoadBackup(path.Join(Config.BackupDir, "backups", backupName))
			// load(server, i)
		}
	} // else if args[0] == "del" {
	// TODO: Backup del
	// for i, index := range(args[1:]) {

	// }
	//}
	return nil
}

func start(server *Server, args... string) error {
	if server.Running {
		return nil
	} else {
		if _, err := os.Stat(server.config.Dir); err != nil {
			log.Panicln(err)
		} else {
			needInstall := false
			downloadServer := false
			if _, err := os.Stat(path.Join(server.config.Dir, "server.jar")); err != nil {
				log.Printf("[%v]: server.jar not found.", server.config)
				needInstall = true
				downloadServer = true
			}
			if _, err := os.Stat(path.Join(server.config.Dir, "quilt-server-launch.jar")); err != nil {
				log.Printf("[%v]: quilt-server-launch.jar not found.", server.config)
				needInstall = true
			}
			if _, err := os.Stat(path.Join(server.config.Dir, "versions", server.config.Version)); err != nil {
				log.Printf("[%v]: version cannot match.", server.Name)
				needInstall = true
				downloadServer = true
			}
			downloadServer = true
			
			if needInstall {
				utils.DeletePath(path.Join(server.config.Dir, ".quilt"))
				utils.DeletePath(path.Join(server.config.Dir, "config"))
				utils.DeletePath(path.Join(server.config.Dir, "libraries"))
				utils.DeletePath(path.Join(server.config.Dir, "versions"))
				utils.DeletePath(path.Join(server.config.Dir, "quilt-server-launch.jar"))
				utils.DeletePath(path.Join(server.config.Dir, "server.jar"))
				if _, err := os.Stat("./quilt-installer-0.5.0.jar"); err != nil {
					log.Printf("quilt-installer-0.5.0.jar not found, downloading...")
					DownloadQuiltInstaller()
				}

				log.Printf("Installing server using quilt-installer-0.5.0.jar...")
				// java -jar quilt-installer-0.5.0.jar install server 1.19.4 --download-server
				args := []string{"-jar", "quilt-installer-0.5.0.jar", "install", "server"}
				if len(server.config.Version) != 0 {
					args = append(args, server.config.Version)
				} else {
					server.config.Version = "1.19.4"
					args = append(args, "1.19.4")
					Config.Servers[server.Name] = server.config
					config.Save(Config)
				}
				if downloadServer {
					args = append(args, "--download-server")
				}
				args = append(args, fmt.Sprintf("--install-dir=%v", server.config.Dir))
				output, err := exec.Command("java", args...).Output()
				if err != nil {
					log.Println(err)
					log.Panicf("Install failed: %v\n%v\n", err, string(output))
				} else {
					log.Println(string(output))
				}
			}
		}
		go server.Run()
	}

	return nil
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

func restart(server *Server, args... string) error {
	server.Write("stop")
	for server.Running {
		time.Sleep(time.Second)
	}
	go server.Run()
	return nil
}
