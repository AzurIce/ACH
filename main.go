package main

import (
	// "ach/bootstrap"
	"ach/bootstrap"
	"ach/core"
	"ach/models"
	"os"

	// "sync"
	"ach/routers"
	// "log"
)


func init() {
	bootstrap.InitStaticFS()
	bootstrap.InitConfig()
	os.Mkdir(bootstrap.Config.BackupDir, 0666)
	models.Init()
	core.Init()
}

func main() {
	// ach := core.Ach()

	api := routers.InitRouter()

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

