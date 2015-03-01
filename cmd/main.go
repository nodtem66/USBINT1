package main

import (
	"flag"
	"fmt"
	"github.com/nodtem66/usbint1/env"
	"os"
	"strings"
)

var (
	deviceString = flag.String("dev", "0000:0000", "device to listen "+
		"VENDOR:PRODUCT (hex) eg. 10C4:FFFF")
)

func main() {
	//get last execute name
	i := strings.LastIndex(os.Args[0], env.PATH_SEPERATOR)
	programName := os.Args[0][i+1:]

	//replace default usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", programName)
		flag.PrintDefaults()
	}
	flag.Parse()
}
