package core

import (
	"log"
	"regexp"
)

var ForwardReg *regexp.Regexp

// var commandReg *regexp.Regexp
var PlayerOutputReg *regexp.Regexp

var OutputFormatReg *regexp.Regexp

func init() {
	log.Println("MCSH[init/INFO]: Initializing regexps...")
	ForwardReg = regexp.MustCompile(`(.+?) *\| *(.+)`)
	// commandReg = regexp.MustCompile("^" + string(MCSHConfig.CommandPrefix) + "(.*)")
	PlayerOutputReg = regexp.MustCompile(`\]: <(.*?)> (.*)`)
	OutputFormatReg = regexp.MustCompile(`(\[\d\d:\d\d:\d\d\]) *\[.+?\/(.+?)\]`)
}
