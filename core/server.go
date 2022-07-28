package core

import (
	"ach/bootstrap"
	"ach/config"
	"ach/lib/utils"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Server ...
type Server struct {
	Name                     string
	config                   config.ServerConfig
	Running                  bool
	InChan, OutChan, ErrChan chan string
	cmdChan                  chan string
	stdin                    io.WriteCloser
	stdout, stderr           io.ReadCloser
	cmd                      *exec.Cmd
}

// NewServer ...
func NewServer(name string, config config.ServerConfig) *Server {
	server := &Server{
		Name:    name,
		config:  config,
		Running: false,
		InChan:  make(chan string, 8),
		OutChan: make(chan string, 8),
		ErrChan: make(chan string, 8),
		cmdChan: make(chan string, 8),
	}

	go server.tick()

	return server
}

func (server *Server) tick() {
	// log.Println("ticking...")
	for {
		// log.Println("Selecting...")
		select {
		case line := <-server.cmdChan:
			// log.Println(line)
			words := strings.Split(line, " ")
			args := []string{""}
			if len(words) > 1 {
				args = words[1:]
			}
			var cmdFun, exist = Cmds[words[0]]
			if exist {
				cmdFun(server, args)
			}
		case line := <-server.InChan:
			if line[:1] == bootstrap.Config.CommandPrefix {
				// log.Println(line)
				server.cmdChan <- line[1:]
			} else if server.Running {
				server.stdin.Write([]byte(line + "\n"))
			}
		case line := <-server.OutChan:
			// 去掉换行符
			if i := strings.LastIndex(string(line), "\r"); i > 0 {
				line = line[:i]
			} else {
				line = line[:len(line)-1]
			}
			if res := PlayerOutputReg.FindStringSubmatch(line); len(res) > 1 { // Player
				// player := res[1]
				text := res[2]
				// log.Println(player + ": " + text)
				if text[:1] == bootstrap.Config.CommandPrefix {
					server.cmdChan <- text[1:]
				}
			}
			str := OutputFormatReg.ReplaceAllString(line, utils.GetTime()+" ["+server.Name+"/$2]") // 格式化读入的字符串
			ACH.Writeln(str)
		case line := <-server.ErrChan:
			log.Print(line)
		}
	}
}

func (server *Server) Write(str string) {
	server.stdin.Write([]byte(str + "\n"))
}

var ErrServerIsRunning = errors.New("Server is already running")

func (server *Server) Run() {
	if err := server.start(); err != nil {
		log.Printf("server<%s>: Error when starting:\n%s\n", server.Name, err)
		return
	}

	if err := server.wait(); err != nil {
		log.Printf("server<%s>: Error when waiting:\n%s\n", server.Name, err)
	}
}

// If successful started, return nil.
func (server *Server) start() error {
	if server.Running {
		return ErrServerIsRunning
	}
	server.initCmd()
	// Start
	server.attachStd()
	if err := server.cmd.Start(); err != nil {
		return err
	}
	server.Running = true
	go utils.ForwardStd(server.stdout, server.OutChan)
	go utils.ForwardStd(server.stderr, server.ErrChan)
	return nil
}

// If successfuly ended, return nil.
func (server *Server) wait() error {
	if err := server.cmd.Wait(); err != nil {
		server.Running = false
		return err
	}
	server.Running = false
	return nil
}

func (server *Server) initCmd() {
	args := append(strings.Split(server.config.ExecOptions, " "), "-jar",
		server.config.ExecPath, "--nogui")
	cmd := exec.Command("java", args...)
	cmd.Dir = filepath.Dir(server.config.ExecPath)

	server.cmd = cmd
}

func (server *Server) attachStd() {
	log.Println("[server/attachStd]: attaching stdin, stdout and stderr...")
	server.stdin, _ = server.cmd.StdinPipe()
	server.stdout, _ = server.cmd.StdoutPipe()
	server.stderr, _ = server.cmd.StderrPipe()
}

// ---- commands ----

func (server *Server) GetBackupList(dst string) []string {
	backupList := make([]string, 0, 10)

	res, _ := ioutil.ReadDir(dst)
	sort.Slice(res, func(i, j int) bool {
		return res[i].ModTime().Unix() < res[j].ModTime().Unix()
	})
	for _, f := range res {
		backupList = append(backupList, f.Name())
	}
	return backupList
}

func (server *Server) MakeBackup(dst string) error {
	src := path.Join(filepath.Dir(server.config.ExecPath), "world")
	log.Printf("[%s/INFO]: Making backup to %s...\n", server.Name, dst)
	server.Write(fmt.Sprintf("say Making backup to %s...", dst))

	server.Write("save-off")
	err := utils.CopyDir(src, dst)
	server.Write("save-on")

	if err != nil {
		log.Printf("[%s/ERROR]: Backup making failed.\n", server.Name)
		server.Write("say Backup making failed.")
		return err
	}
	log.Printf("[%s/INFO]: Backup making succeed.\n", server.Name)
	server.Write("say Backup making succeed.")
	return nil
}

func (server *Server) LoadBackup(src string) error {
	dst := path.Join(filepath.Dir(server.config.ExecPath), "world")
	server.MakeSnapshot("Before loading backup")

	server.Write("stop")
	for server.Running {
		time.Sleep(time.Second)
	}

	utils.DeletePath(dst)

	log.Printf("[%s/INFO]: Loading backup %s...\n", server.Name, src)
	err := utils.CopyDir(src, dst)
	if err != nil {
		log.Printf("[%s/ERROR]: Backup loading failed.\n", server.Name)
		return err
	}
	log.Printf("[%s/INFO]: Backup loaded successfully.\n", server.Name)

	server.start()
	go server.wait()

	return nil
}

func (server *Server) GetSnapshotCount() int {
	res, _ := ioutil.ReadDir(path.Join(bootstrap.Config.BackupDir, "snapshots"))
	return len(res)
}

func (server *Server) GetSnapshotList() []string {
	return server.GetBackupList(path.Join(bootstrap.Config.BackupDir, "snapshots"))
}

func (server *Server) DeleteOldSnapshot() {
	res, _ := ioutil.ReadDir(path.Join(bootstrap.Config.BackupDir, "snapshots"))
	sort.Slice(res, func(i, j int) bool {
		return res[i].ModTime().Unix() < res[j].ModTime().Unix()
	})

	utils.DeletePath(path.Join(bootstrap.Config.BackupDir, "snapshots", res[0].Name()))
}

func (server *Server) MakeSnapshot(comment string) error {
	for server.GetSnapshotCount() >= 10 {
		server.DeleteOldSnapshot()
	}

	bkname := utils.GetTimeStamp()
	if len(comment) > 0 {
		bkname += " " + comment
	}

	if err := server.MakeBackup(
		path.Join(
			bootstrap.Config.BackupDir,
			"snapshots",
			fmt.Sprintf("%s - %s", server.Name, bkname),
		),
	); err != nil {
		return err
	}
	return nil
}

func (server *Server) LoadSnapshot(index int) error {
	snapshotName := server.GetSnapshotList()[index]
	snapshotPath := path.Join(bootstrap.Config.BackupDir, "snapshots", snapshotName)
	if err := server.LoadBackup(snapshotPath); err != nil {
		return err
	}
	return nil
}
