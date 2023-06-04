package service

import (
	"ach/core"
	"ach/internal/models"
	"log"
    "io"

	"github.com/gin-gonic/gin"
)

type GetServersService struct{}

func (s *GetServersService) Handle(c *gin.Context) (any, error) {
	servers := models.GetServers()
	log.Println("[services/server/GetServers]: ", servers)
	// bytes, _ := json.Marshal(servers)
	return servers, nil
}

func ServerConsoleHandler() gin.HandlerFunc {
    return func (c *gin.Context) {
        log.Println("ServerConsole")
        c.Writer.Header().Set("Content-Type", "text/event-stream")
        c.Writer.Header().Set("Cache-Control", "no-cache")

        messageChan := make(chan string)

        flag := true
        done := c.Stream(func(w io.Writer) bool {
            if flag {
                c.SSEvent("message", core.ACH.OutBuf.GetBuf())
                core.ACH.AddSSEChan(&messageChan)
                flag = false
                return true
            }
            select {
            case message := <- messageChan:
                c.SSEvent("message", message)
            }
            return true
        })

        if done {
            core.ACH.RemoveSSEChan(&messageChan)
        }
    }
}

type ServerConsolePostService struct{
	Data string `form:"data"`
}

func (s *ServerConsolePostService) Handle(c *gin.Context) (any, error) {
	core.ACH.InChan <- string(s.Data)
	return nil, nil
}