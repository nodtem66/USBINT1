package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	DEBUG     bool = true
	LOG_LEVEL int  = 4
)

func GetProgramName(cmd string) (programName string) {
	i := strings.LastIndex(cmd, PATH_SEPERATOR)
	programName = cmd[i+1:]
	return
}
func GetVidPidFromString(str string) (vid int, pid int, err error) {
	// parse XXX:XXX to device vendorId and productId
	// split AAA:BBB to [AAA BBB]
	substringDevice := strings.Split(str, ":")

	// If length of substring is not 2, print error
	if len(substringDevice) != 2 {
		err = fmt.Errorf("%s is not a valid parameter", str)
	}
	// convert string to hex
	if hex, err := strconv.ParseUint(substringDevice[0], 16, 16); err == nil {
		vid = (int)(hex)
	}
	if hex, err := strconv.ParseUint(substringDevice[1], 16, 16); err == nil {
		pid = (int)(hex)
	}
	return
}

func GetHostPortFromString(str string) (host string, port int, err error) {
	params := strings.Split(str, ":")
	host = params[0]
	if len(params) == 1 {
		port = 8086
	} else if len(params) == 2 {
		port, err = strconv.Atoi(params[1])
	} else {
		err = fmt.Errorf("%s is not a valid host:port for influxdb", str)
	}
	return
}
