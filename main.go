package main

import (
	"ach/core"
)

func main() {
	ach := core.Ach()

	// ach.TestRun()
	// ach.Run()
	ach.TestRouter()
}
