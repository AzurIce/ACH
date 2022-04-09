package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ach/bootstrap"
	"ach/utils"
)

// Command ...
type Command struct {
	Cmd  string
	Args []string
}

// Cmds ...
var Cmds = make(map[string]interface{})

func backup(server *Server, args []string) error {
	if args[0] == "make" {
		comment := utils.GetTimeStamp()
		if len(args) > 1 {
			comment = comment + " " + strings.Join(args[1:], " ")
		}
		dst := path.Join(bootstrap.Config.BackupDir, fmt.Sprintf("%s - %s", server.name, comment))
		src := path.Join(filepath.Dir(server.config.ExecPath), "world")
		log.Printf("[%s/INFO]: Making backup to %s...\n", server.name, dst)
		server.Write(fmt.Sprintf("say Making backup to %s...", dst))
		err := utils.CopyDir(src, dst)
		if err != nil {
			log.Printf("[%s/ERROR]: Backup making failed.\n", server.name)
			server.Write("say Backup making failed.")
			return err
		}
		log.Printf("[%s/INFO]: Backup making successed.\n", server.name)
		server.Write("say Backup making successed.")
	} else if args[0] == "" || args[0] == "list" {
		// log.Printf("[%s/INFO]: Listing backup.\n", server.ServerName)
		res, _ := ioutil.ReadDir(bootstrap.Config.BackupDir)
		for i, f := range res {
			fmt.Printf("[%v] %s\n", i, f.Name())
			server.Write(fmt.Sprintf("say [%v] %s", i, f.Name()))
		}
	} else if args[0] == "load" {
		i, err := strconv.Atoi(strings.Join(args[1:], ""))
		if err == nil {
			load(server, i)
		}
	} else if args[0] == "del" {
		// TODO: Backup del
		// for i, index := range(args[1:]) {

		// }
	}
	return nil
}

func load(server *Server, i int) error {
	// TODO: Err handle
	res, _ := ioutil.ReadDir(bootstrap.Config.BackupDir)
	backup(server, []string{"make", fmt.Sprintf("Before loading %s", res[i].Name())})

	server.keepAlive = true
	server.Write("stop")
	for server.running {
		time.Sleep(time.Second)
	}

	backupSavePath := path.Join(bootstrap.Config.BackupDir, res[i].Name())
	serverSavePath := path.Join(filepath.Dir(server.config.ExecPath), "world")
	os.RemoveAll(serverSavePath)

	log.Printf("[%s/INFO]: Loading backup %s...\n", server.name, res[i].Name())
	err := utils.CopyDir(backupSavePath, serverSavePath)
	if err != nil {
		log.Printf("[%s/ERROR]: Backup loading failed.\n", server.name)
		server.keepAlive = false
		return err
	}
	log.Printf("[%s/INFO]: Backup loading successed.\n", server.name)

	server.Start()
	server.keepAlive = false
	go server.Wait()

	return nil
}

func start(server *Server, args []string) error {
	if server.running {
		return nil
	} else {
		server.Start()
		go server.Wait()
	}

	return nil
}

func restart(server *Server, args []string) error {
	server.keepAlive = true
	server.Write("stop")
	for server.running {
		time.Sleep(time.Second)
	}
	server.Start()
	go server.Wait()
	return nil
}

func init() {
	log.Println("MCSH[init/INFO]: Initializing commands...")
	Cmds["backup"] = backup
	Cmds["start"] = start
	Cmds["restart"] = restart
	// Cmds["clone"] = clone
}
