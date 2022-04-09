package main

import (
	"ach/core"
	"ach/routers"
)

func main() {
	// ach := core.Ach()
	core.Init()

	api := routers.InitRouter()

	api.Run()

	// ach.TestRun()
	// ach.Run()
	// ach.TestRouter()
}
