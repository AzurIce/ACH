package bootstrap

import "flag"

var Dev bool

func init() {
    flag.BoolVar(&Dev, "dev", false, "Whether to use dev server")
    flag.Parse()
}


