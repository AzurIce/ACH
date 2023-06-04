package core

import (
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"ach/internal/bootstrap"
	"ach/internal/utils"
)


var Commands = map[string] interface{} {}

func RegisterCommand(command string, f func() interface{}) {
	
}

type Command interface {
	Command() string
}

type ACHCommand interface {
	Command
	Execute(server *Server, args []string) error
}
type ServerCommand interface {
	Command
	Execute(args []string) error
}

// Command ...
// type Command struct {
// 	Cmd  string
// 	Args []string
// }

// Cmds ...
var Cmds = map[string]interface{}{
	"backup":  back,
	"bksnap":  &bkSnapServerCommand{},
	"start":   start,
	"restart": restart,
}

func init() {
	Cmds[bkSnapServerCommand.Command()] = &bkSnapServerCommand{}
}

// The length of the "args" argument is >= 1, if no args input, it will be [""]
type bkSnapServerCommand struct{}
func (c *bkSnapServerCommand) Command() string {
	return "#bksnap"
}
func (c *bkSnapServerCommand) Execute(server *Server, args []string) error {
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

type backupServerCommand struct{}
func (c *backupServerCommand)Execute(server *Server, args []string) error {
	if args[0] == "make" {
		comment := utils.GetTimeStamp()
		if len(args) > 1 {
			comment += " " + strings.Join(args[1:], " ")
		}
		dst := path.Join(bootstrap.Config.BackupDir, "backups", fmt.Sprintf("%s - %s", server.Name, comment))
		if err := server.MakeBackup(dst); err != nil {
			log.Println(err)
			return err
		}
	} else if args[0] == "" || args[0] == "list" {
		backupList := server.GetBackupList(path.Join(bootstrap.Config.BackupDir, "backups"))
		// log.Printf("[%s/INFO]: Listing backup.\n", server.ServerName)
		for i, backup := range backupList {
			fmt.Printf("[%v] %s\n", i, backup.FileName)
			server.Write(fmt.Sprintf("say [%v] %s", i, backup.FileName))
		}
	} else if args[0] == "load" {
		i, err := strconv.Atoi(strings.Join(args[1:], ""))
		if err == nil {
			backupName := server.GetBackupList(path.Join(bootstrap.Config.BackupDir, "backups"))[i].FileName
			server.LoadBackup(path.Join(bootstrap.Config.BackupDir, "backups", backupName))
			// load(server, i)
		}
	} // else if args[0] == "del" {
	// TODO: Backup del
	// for i, index := range(args[1:]) {

	// }
	//}
	return nil
}

type startACHCommand struct{}
func (c *startACHCommand)Execute(args []string) error {
	if server.Running {
		return nil
	} else {
		go server.Run()
	}

	return nil
}

type restartACHCommand struct{}
func (c *restartACHCommand)Execute(args []string) error {
	server.Write("stop")
	for server.Running {
		time.Sleep(time.Second)
	}
	go server.Run()
	return nil
}

type restartServerCommand struct{}
func (c *restartServerCommand)Execute(server *Server, args []string) error {
	server.Write("stop")
	for server.Running {
		time.Sleep(time.Second)
	}
	go server.Run()
	return nil
}
