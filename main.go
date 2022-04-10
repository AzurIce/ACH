package main

import (
	"ach/core"
	"ach/routers"
)

func main() {
	// ach := core.Ach()
	core.Init()

	api := routers.InitRouter()

	api.Run(":8888")

	// ach.TestRun()
	// ach.Run()
	// ach.TestRouter()
}
