package main

import (
	"ach/core"
)

func main() {
	ach := core.Ach()
	ach.Run()
	// ach.TestRouter()
}
