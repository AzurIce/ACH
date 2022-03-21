package main

// import "time"

import (
	"ach/core"
)


func main() {
	ach := core.Ach()
	ach.Run()
	// ach.TestRun()
}

// func processInput(ach *core.ACHCore, reader *bufio.Reader) {
// 	stdinReader := bufio.NewReader(os.Stdin)
// 	for {
// 		line, errRead := stdinReader.ReadBytes('\n')
// 		if errRead != nil {
// 			log.Println("MCSH[stdinForward/ERROR]: ", errRead)
// 		} else {
// 			// 去掉换行符
// 			if i := strings.LastIndex(string(line), "\r"); i > 0 {
// 				line = line[:i]
// 			} else {
// 				line = line[:len(line)-1]
// 			}
// 			// 转发正则
// 			res := core.ForwardReg.FindSubmatch(line)
// 			if res != nil { // 转发到特定服务器
// 				server, exist := ach.Servers[string(res[1])]
// 				if exist {
// 					server.InChan <- string(res[2])
// 				} else {
// 					log.Printf("MCSH[stdinForward/ERROR]: Cannot find running server <%v>\n", string(res[1]))
// 				}
// 			} else { // 转发到所有服务器
// 				for _, server := range ach.Servers {
// 					server.InChan <- string(line)
// 				}
// 			}
// 		}
// 	}
// }
