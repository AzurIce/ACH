package core

import (
	"ach/utils"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// Server ...
type Server struct {
	ach                      *ACHCore
	name                     string
	config                   ServerConfig
	running                  bool
	keepAlive                bool
	InChan, OutChan, ErrChan chan string
	cmdChan                  chan string
	stdin                    io.WriteCloser
	cmd                      *exec.Cmd
}

// NewServer ...
func NewServer(name string, config ServerConfig, ach *ACHCore) *Server {
	server := &Server{
		ach:       ach,
		name:      name,
		config:    config,
		running:   false,
		keepAlive: false,
		InChan:    make(chan string, 8),
		OutChan:   make(chan string, 8),
		ErrChan:   make(chan string, 8),
		cmdChan:   make(chan string),
	}

	go server.processIn()
	go server.handleCommand()

	return server
}

func (server *Server) Write(str string) {
	server.stdin.Write([]byte(str + "\n"))
}

func (server *Server) initCmd() {
	args := append(strings.Split(server.config.ExecOptions, " "), "-jar",
		server.config.ExecPath, "--nogui")
	cmd := exec.Command("java", args...)
	cmd.Dir = filepath.Dir(server.config.ExecPath)

	server.cmd = cmd
}

// Start ...
func (server *Server) Start() {
	defer func() {
		recover()
	}()
	server.initCmd()
	server.attachStd()
	// Start
	if !server.keepAlive {
		server.ach.wg.Add(1)
	}
	server.running = true
	if err := server.cmd.Start(); err != nil {
		log.Panicf("server<%s>: Error when starting:\n%s", server.name, err.Error())
	}
}

// Wait...
func (server *Server) Wait() {
	defer func() {
		recover()
	}()
	if err := server.cmd.Wait(); err != nil {
		log.Panicf("server<%s>: Error when running:\n%s", server.name, err.Error())
	}
	server.running = false
	if !server.keepAlive {
		server.ach.wg.Done()
	}
}

func (server *Server) attachStd() {
	server.stdin, _ = server.cmd.StdinPipe()
	stdout, _ := server.cmd.StdoutPipe()
	go utils.ForwardStd(stdout, server.OutChan)
	go server.processOut()
	stderr, _ := server.cmd.StderrPipe()
	go utils.ForwardStd(stderr, server.ErrChan)
	go server.processErr()
}

func (server *Server) handleCommand() {
	for {
		line := <-server.cmdChan
		words := strings.Split(line, " ")
		args := []string{""}
		if len(words) > 1 {
			args = words[1:]
		}
		var cmdFun, exist = Cmds[words[0]]
		if exist {
			cmdFun.(func(server *Server, args []string) error)(server, args)
		}
	}
}
func (server *Server) processIn() {
	for {
		line := <-server.InChan
		if line[:1] == server.ach.config.CommandPrefix {
			server.cmdChan <- line[1:]
		} else if server.running {
			server.stdin.Write([]byte(line + "\n"))
		}
	}
}
func (server *Server) processOut() {
	for server.running {
		line := <-server.OutChan
		// 去掉换行符
		if i := strings.LastIndex(string(line), "\r"); i > 0 {
			line = line[:i]
		} else {
			line = line[:len(line)-1]
		}
		if res := playerOutputReg.FindStringSubmatch(line); len(res) > 1 { // Player
			player := res[1]
			text := res[2]
			log.Println(player + ": " + text)
			if text[:1] == server.ach.config.CommandPrefix {
				server.cmdChan <- text[1:]
			}
		}
		str := outputFormatReg.ReplaceAllString(line, "["+server.name+"/$2]") // 格式化读入的字符串
		server.ach.println(str)
		// server.ach.OutChan <- str // 嘿嘿！
		// log.Print(str)
	}
}
func (server *Server) processErr() {
	for server.running {
		line := <-server.ErrChan
		log.Print(line)
	}
}
