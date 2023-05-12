package main

import (
	"ach/core"
	"ach/server"
	"log"
)

func init() {
	core.Init()
}

func main() {
	api := server.InitRouter()

	err := api.Run(":8888")
	if err != nil {
		log.Panicln(err)
	}
}
