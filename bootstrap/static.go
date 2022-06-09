package bootstrap

import (
	"log"
	"net/http"

	_ "ach/statik"
	"github.com/rakyll/statik/fs"
)

var StaticFS http.FileSystem

func InitStaticFS() {
	
	var err error
	StaticFS, err = fs.New()
	if err != nil {
		log.Panicln(err)
	}
}