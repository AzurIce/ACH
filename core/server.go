package core

import (
	"ach/bootstrap"
	"ach/config"
	"ach/lib/utils"
	"errors"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// Server ...
type Server struct {
	name                     string
	config                   config.ServerConfig
	running                  bool
	InChan, OutChan, ErrChan chan string
	cmdChan                  chan string
	stdin                    io.WriteCloser
	stdout, stderr           io.ReadCloser
	cmd                      *exec.Cmd
}

// NewServer ...
func NewServer(name string, config config.ServerConfig) *Server {
	server := &Server{
		name:    name,
		config:  config,
		running: false,
		InChan:  make(chan string, 8),
		OutChan: make(chan string, 8),
		ErrChan: make(chan string, 8),
		cmdChan: make(chan string),
	}

	go server.tick()
	// go server.processIn()
	// go server.handleCommand()

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
				cmdFun.(func(server *Server, args []string) error)(server, args)
			}
		case line := <-server.InChan:
			if line[:1] == bootstrap.Config.CommandPrefix {
				server.cmdChan <- line[1:]
			} else if server.running {
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
			str := OutputFormatReg.ReplaceAllString(line, "["+server.name+"/$2]") // 格式化读入的字符串
			ACH.Println(str)
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
		log.Printf("server<%s>: Error when starting:\n%s\n", server.name, err)
		return
	}

	if err := server.wait(); err != nil {
		log.Printf("server<%s>: Error when waiting:\n%s\n", server.name, err)
	}
}


// If successful started, return nil.
func (server *Server) start() error {
	if server.running {
		return ErrServerIsRunning
	}
	server.initCmd()
	// Start
	server.attachStd()
	if err := server.cmd.Start(); err != nil {
		return err
	}
	server.running = true
	go utils.ForwardStd(server.stdout, server.OutChan)
	go utils.ForwardStd(server.stderr, server.ErrChan)
	return nil
}

// If successfuly ended, return nil.
func (server *Server) wait() error {
	if err := server.cmd.Wait(); err != nil {
		server.running = false
		return err
	}
	server.running = false
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

// func (server *Server) handleCommand() {
// 	for {
// 		line := <-server.cmdChan
// 		words := strings.Split(line, " ")
// 		args := []string{""}
// 		if len(words) > 1 {
// 			args = words[1:]
// 		}
// 		var cmdFun, exist = Cmds[words[0]]
// 		if exist {
// 			cmdFun.(func(server *Server, args []string) error)(server, args)
// 		}
// 	}
// }

// func (server *Server) processIn() {
// 	for {
// 		line := <-server.InChan
// 		if line[:1] == bootstrap.Config.CommandPrefix {
// 			server.cmdChan <- line[1:]
// 		} else if server.running {
// 			server.stdin.Write([]byte(line + "\n"))
// 		}
// 	}
// }

// func (server *Server) processOut() {
// 	// log.Println("[serve/processOut]: server.running:", server.running)
// 	for server.running {
// 		line := <-server.OutChan
// 		log.Println("[serve/processOut]: line:", line)
// 		// 去掉换行符
// 		if i := strings.LastIndex(string(line), "\r"); i > 0 {
// 			line = line[:i]
// 		} else {
// 			line = line[:len(line)-1]
// 		}
// 		if res := PlayerOutputReg.FindStringSubmatch(line); len(res) > 1 { // Player
// 			// player := res[1]
// 			text := res[2]
// 			// log.Println(player + ": " + text)
// 			if text[:1] == bootstrap.Config.CommandPrefix {
// 				server.cmdChan <- text[1:]
// 			}
// 		}
// 		str := OutputFormatReg.ReplaceAllString(line, "["+server.name+"/$2]") // 格式化读入的字符串
// 		ACH.Println(str)
// 		// log.Println(str)
// 	}
// }
// func (server *Server) processErr() {
// 	for server.running {
// 		line := <-server.ErrChan
// 		log.Print(line)
// 	}
// }
