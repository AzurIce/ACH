package core

import (
	"fmt"
	"log"
	"regexp"

	"ach/internal/config"
	"ach/internal/utils"
	_ "ach/statik"

	"github.com/fsnotify/fsnotify"
)

var ForwardReg = regexp.MustCompile(`(.+?) *\| *(.+)`)

// ACHCore ...
type ACHCore struct {
	Servers map[string]*Server

	LogChan   chan string
	LogBuf    *utils.ScrollBuffer

	OutChan   chan string
	OutBuf    *utils.ScrollBuffer

	SSEChanList []*chan string

	InChan chan string
}

var ACH *ACHCore

func init() {
	ACH = &ACHCore{
		Servers: make(map[string]*Server),
		InChan:  make(chan string, 8),

		OutChan:   make(chan string, 8),
		OutBuf:    utils.NewScrollBuffer(),

		LogChan:   make(chan string, 8),
		LogBuf:    utils.NewScrollBuffer(),
	}

    ACH.UpdateServers(*Config)

	go ACH.tick()
}

func (ach *ACHCore) UpdateServers(config config.ACHConfig) {
	for name, serverConfig := range config.Servers {
        if server, ok := ach.Servers[name]; ok {
		    ach.Servers[name].config = serverConfig
            if server.Running {
                if server.config.Dir != config.Servers[name].Dir ||
                   server.config.Version != config.Servers[name].Version {
                    restart(ach.Servers[name])
                }
            }
        } else {
		    ACH.Servers[name] = NewServer(name, serverConfig)
        }
	}
}

func (ach *ACHCore) StartAllServers() {
	log.Println("[ACHCore]: Starting all servers")
	for _, server := range ach.Servers {
		go server.Run()
	}
}

func (ach *ACHCore) tick() {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatalln(err)
    }
    defer watcher.Close()

    // Start listening for events.
    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                // log.Println("event:", event)
                if event.Has(fsnotify.Write) {
                    log.Println("Detected config.yml changed: ", event.Name)
                    Config, err = config.ReadConfig()
                    if err != nil {
                        log.Println("Failed to read config.yml: ", err)
                    }
                    ach.UpdateServers(*Config)
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("error:", err)
            }
        }
    }()

    // Add a path.
    err = watcher.Add("./config.yml")
    if err != nil {
        log.Fatal(err)
    }

	for {
		select {
		case line := <-ach.InChan:
			// 转发正则
			res := ForwardReg.FindStringSubmatch(line)
			if res != nil { // 转发到特定服务器
				server, exist := ach.Servers[string(res[1])]
				if exist {
					server.InChan <- string(res[2])
				} else {
					log.Printf("[ACHCore/tick]: Cannot find running server <%v>\n", string(res[1]))
				}
			} else { // 转发到所有服务器
				for _, server := range ach.Servers {
					server.InChan <- string(line)
				}
			}
		case line := <-ach.OutChan:
			fmt.Print("[ACHCore/tick]: OutChan: ", line)
			for _, c := range ach.SSEChanList {
				if c != nil {
					*c <- line
				}
			}
			// ach.OutWsPool.AllSendMessage(line)
		case line := <-ach.LogChan:
			log.Print("[ACHCore/tick]: LogChan: ", line)
			for _, c := range ach.SSEChanList {
				if c != nil {
					*c <- line
				}
			}
			// ach.LogWsPool.AllSendMessage(line)
		}
	}
}

func (ach *ACHCore) AddSSEChan(ch *chan string) {
	for i, ws := range ach.SSEChanList {
		if ws == nil {
			ach.SSEChanList[i] = ch
			return
		}
	}
	ach.SSEChanList = append(ach.SSEChanList, ch)
}

func (ach *ACHCore) RemoveSSEChan(ch *chan string) {
    for i, channel := range ach.SSEChanList {
        if channel == ch {
            ach.SSEChanList[i] = nil
        }
    }
}

// --- init ---

func (ach *ACHCore) Log(str string) {
	ach.LogBuf.Write(str)
	ach.LogChan <- str
}

func (ach *ACHCore) Logln(str string) {
	ach.Log(str + "\n")
}

func (ach *ACHCore) Write(str string) {
	ach.OutBuf.Write(str)
	ach.OutChan <- str
}

func (ach *ACHCore) Writeln(str string) {
	ach.Write(str + "\n")
}
