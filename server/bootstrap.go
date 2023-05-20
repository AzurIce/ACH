package server

import (
    "log"
	"net/http"
	"github.com/rakyll/statik/fs"
)

// StaticFS
var StaticFS http.FileSystem

func init() {
	log.Println("[bootStrap/InitStaticFS]: Initializing...")
	var err error
	StaticFS, err = fs.New()
	if err != nil {
		log.Panicln(err)
	}
}
