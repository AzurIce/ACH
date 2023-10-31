package bootstrap

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)
type FS struct {
	FS http.FileSystem
}

var StaticFS *FS

func InitStatic(statics embed.FS) {
	log.Println("[bootStrap/InitStaticFS]: Initializing...")
	var err error
	embedFS, err := fs.Sub(statics, "assets/dist")
	if err != nil {
		log.Panicf("Failed to initialize static resources: %s", err)
	}

	StaticFS = &FS {
		http.FS(embedFS),
	}
}


// Open 打开文件
func (b *FS) Open(name string) (http.File, error) {
	return b.FS.Open(name)
}

// Exists 文件是否存在
func (b *FS) Exists(prefix string, filepath string) bool {
	if _, err := b.FS.Open(filepath); err != nil {
		return false
	}
	return true
}