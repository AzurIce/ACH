package core

import (
	"log"
	"regexp"
)

var ForwardReg *regexp.Regexp

// var commandReg *regexp.Regexp
var playerOutputReg *regexp.Regexp

var outputFormatReg *regexp.Regexp

func init() {
	log.Println("MCSH[init/INFO]: Initializing regexps...")
	ForwardReg = regexp.MustCompile(`(.+?) *\| *(.+)`)
	// commandReg = regexp.MustCompile("^" + string(MCSHConfig.CommandPrefix) + "(.*)")
	playerOutputReg = regexp.MustCompile(`\]: <(.*?)> (.*)`)
	outputFormatReg = regexp.MustCompile(`(\[\d\d:\d\d:\d\d\]) *\[.+?\/(.+?)\]`)
}
