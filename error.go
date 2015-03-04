package usbint

import (
	"fmt"
)

type usbintError int

func (e usbintError) Error() string {
	return fmt.Sprintf("usbint1: [%d] %s ", int(e), usbintErrorString[e])
}

const (
	SUCCESS usbintError = iota
)

var usbintErrorString = map[usbintError]string{
	SUCCESS: "success",
}
