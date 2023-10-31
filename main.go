package main

import (
	"ach/core"
	"ach/internal/bootstrap"
	"ach/server"
	"embed"
	"log"
)

//go:embed all:assets/dist/*
var f embed.FS

func init() {
	bootstrap.InitStatic(f)
	core.Init()
}

func main() {
	api := server.InitRouter()

	err := api.Run(":8888")
	if err != nil {
		log.Panicln(err)
	}
}
