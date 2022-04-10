package main

import (
	"ach/bootstrap"
	"ach/core"
	"ach/routers"
	"log"
)

func main() {
	// ach := core.Ach()
	core.Init()

	api := routers.InitRouter()

	log.Print("[main]: config: ", bootstrap.Config, '\n')

	core.ACH.StartAllServers()

	api.Run(":8888")

	// ach.TestRun()
	// ach.Run()
	// ach.TestRouter()
}
