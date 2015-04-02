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
