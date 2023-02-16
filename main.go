package main

import (
	// "ach/bootstrap"
	// "ach/bootstrap"
	"ach/core"
	// "ach/models"

	// "sync"
	"ach/server"
	// "log"
)

func init() {
	// bootstrap.InitStaticFS()
	// bootstrap.InitConfig()
	// bootstrap.InitDirs()
	// models.Init()
	core.Init()
}

func main() {
	// ach := core.Ach()

	api := server.InitRouter()

	// log.Print("[main]: config: ", bootstrap.Config, '\n')

	// var wg sync.WaitGroup
	// wg.Add(1)

	// core.ACH.StartAllServers()
	// wg.Wait()

	api.Run(":8888")

	// ach.TestRun()
	// ach.Run()
	// ach.TestRouter()
}
