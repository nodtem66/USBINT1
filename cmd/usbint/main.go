package main

import (
	"flag"
	"fmt"
	"github.com/nodtem66/usbint1"
	"github.com/nodtem66/usbint1/db"
	"github.com/nodtem66/usbint1/firmware"
	"os"
	"os/signal"
	"regexp"
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
	NewDB     bool
}

func (c *CommandLine) ParseOption() (err error) {

	// check there is patientID
	if len(c.PatientId) == 0 {
		err = fmt.Errorf("`%s` is not a valid patient id", c.PatientId)
		return
	}
	var match bool
	if match, err = regexp.MatchString("^[0-9a-zA-Z]+$", c.PatientId); match != true {
		if err == nil {
			err = fmt.Errorf("`%s` is not a valid patien id", c.PatientId)
		}
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
	fs.BoolVar(&c.NewDB, "n", false, "new sqlite database file")
	fs.Parse(os.Args[1:])

	if err := c.ParseOption(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fs.PrintDefaults()
		return
	}

	// print infomation about daem
	fmt.Printf("[Patiend ID: %s] [USB device %04X:%04X]\n", c.PatientId, c.Vid, c.Pid)

	// start io

	io := usbint.NewIOHandle()
	io.Dev.OpenDevice(c.Vid, c.Pid)
	if io.Dev.OpenErr != nil {
		fmt.Fprintln(os.Stderr, "No devices")
		return
	}

	// init sqlite
	sqlite := db.NewSqliteHandle()
	sqlite.PatientId = c.PatientId

	if c.NewDB {
		if err := sqlite.ConnectNew(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	} else {
		if err := sqlite.Connect(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}

	// start firmware
	f := firmware.NewFirmware(c.PatientId, io, sqlite)

	//start all services
	sqlite.Start()
	io.Start()
	// hook os signal
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt)
	signal.Notify(osSignal, os.Kill)
	done := make(chan bool)
	go func() {
		for sig := range osSignal {
			fmt.Printf("Event: %s\n", sig.String())
			io.Stop()
			done <- true
		}
	}()

	// wait for interrupt
	<-done
	// save database
	sqlite.Stop()
	sqlite.Close()
}
