package core

import (
	"ach/internal/bootstrap"
	"ach/internal/config"
	"ach/internal/utils"
	"errors"
	"io"
	"log"
	"os/exec"
	"path"
	// "path/filepath"
	"regexp"
	"strings"
)

var PlayerOutputReg = regexp.MustCompile(`]: (\[.*?] )?<(.*?)> (.*)`)

var OutputFormatReg = regexp.MustCompile(`(\[\d\d:\d\d:\d\d]) *\[.+?/(.+?)]`)

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
	for {
		select {
		case line := <-server.cmdChan:
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
				// player := res[2]
				text := res[3]
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

// If successfully ended, return nil.
func (server *Server) wait() error {
	if err := server.cmd.Wait(); err != nil {
		server.Running = false
		return err
	}
	server.Running = false
	return nil
}

func (server *Server) initCmd() {
	execFile := "quilt-server-launch.jar"
	// if server.config.LauncherType == "vanilla" {
	// 	execFile = "server.jar"
	// } else if server.config.LauncherType == "fabric" {
	// 	execFile = "fabric-server-launch.jar"
	// } else if server.config.LauncherType == "quilt" {
	// 	execFile = "quilt-server-launch.jar"
	// }
	args := append(strings.Split(server.config.JVMOptions, " "), "-jar",
		path.Join(server.config.Dir,execFile) , "--nogui")
	cmd := exec.Command("java", args...)
	cmd.Dir = server.config.Dir

	server.cmd = cmd
}

func (server *Server) attachStd() {
	log.Println("[server/attachStd]: attaching stdin, stdout and stderr...")
	server.stdin, _ = server.cmd.StdinPipe()
	server.stdout, _ = server.cmd.StdoutPipe()
	server.stderr, _ = server.cmd.StderrPipe()
}
