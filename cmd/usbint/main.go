package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

const (
	DEFAULT_DEVICE_VID_PID = "10C4:8846"
)

type CommandLine struct {
	Vid       int
	Pid       int
	VidPid    string
	PatientId string
	Verbose   bool
}

func (c *CommandLine) ParseOption() (err error) {

	// check there is patientID
	if len(c.PatientId) == 0 {
		err = fmt.Errorf("`%s` is not a valid patient id", c.PatientId)
		return
	}

	// parse XXX:XXX to device vendorId and productId
	// split AAA:BBB to [AAA BBB]
	s := strings.Split(c.VidPid, ":")

	// If length of substring is not 2, print error
	if len(s) != 2 {
		err = fmt.Errorf("`%s` is not a valid vid:pid", c.VidPid)
		return
	}
	// convert string to hex
	var hex uint64
	if hex, err = strconv.ParseUint(s[0], 16, 16); err == nil {
		c.Vid = (int)(hex)
	}
	if hex, err = strconv.ParseUint(s[1], 16, 16); err == nil {
		c.Pid = (int)(hex)
	}

	//logger.SetLogLevel()
	return
}

func main() {
	c := &CommandLine{}

	fs := flag.NewFlagSet("default", flag.ExitOnError)
	fs.StringVar(&c.PatientId, "patient", "", `patient id to store as measurement unit`)
	fs.StringVar(&c.PatientId, "id", "", `patient id (shorthand)`)
	fs.StringVar(&c.VidPid, "dev", DEFAULT_DEVICE_VID_PID, `device to listen in hex format of VENDOR:PRODUCT`)
	fs.StringVar(&c.VidPid, "d", DEFAULT_DEVICE_VID_PID, `device (shorthand)`)
	fs.BoolVar(&c.Verbose, "v", false, "enable verbose mode")
	fs.Parse(os.Args[1:])

	if err := c.ParseOption(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fs.PrintDefaults()
	}

	// print infomation about daem
	fmt.Printf("[Patiend ID: %s] [USB device %04X:%04X]", c.PatientId, c.Vid, c.Pid)

	// hook os signal
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt)
	signal.Notify(osSignal, os.Kill)
	done := make(chan bool)
	go func() {
		for sig := range osSignal {
			fmt.Printf("Event: %s\n", sig.String())
			done <- true
		}
	}()
	<-done
}
