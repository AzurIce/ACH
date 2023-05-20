package main

import (
	"ach/internal/models"
	"ach/server"
	"flag"
	"log"
	"os"
)

func main() {
    resetAdminPassword := flag.Bool("resetAdminPassword", false, "resetAdminPassword")
    flag.Parse()
    if *resetAdminPassword {
        err := models.DeleteUserById(1)
        if err != nil {
            log.Panicln(err)
        }
        err = models.AddDefaultUser()
        if err != nil {
            log.Panicln(err)
        }
        os.Exit(0)
    }

	api := server.InitRouter()

	err := api.Run(":8888")
	if err != nil {
		log.Panicln(err)
	}
}
