package main

import (
	"flag"
	"fmt"
	. "github.com/nodtem66/usbint1"
	"github.com/nodtem66/usbint1/config"
	. "github.com/nodtem66/usbint1/event"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_DEVICE_VID_PID string = "10C4:8846"
	DEFAULT_INFLUXDB_HOST  string = "127.0.0.1:8086"
)

var (
	deviceString = flag.String("dev", DEFAULT_DEVICE_VID_PID, "device to "+
		"listen VENDOR:PRODUCT (hex).")
	hostInfluxDBString = flag.String("influxdb", "", "Influxdb API address "+
		"to host the streaming data.")
)

func main() {
	//get last execute name
	i := strings.LastIndex(os.Args[0], config.PATH_SEPERATOR)
	programName := os.Args[0][i+1:]

	//replace default usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", programName)
		flag.PrintDefaults()
	}
	flag.Parse()

	vid, pid := GetVidPidFromString(*deviceString)
	fmt.Printf("Initialize USB scanner to device %04X:%04X\n", vid, pid)

	// create channel for exit main program
	mainEvent := NewEventSubcriptor()
	event := NewEventHandler()
	event.Start()
	event.Subcribe(EVENT_MAIN, mainEvent)

	// start database interface

	// start scanner
	scanner := NewScanner(vid, pid)
	scanner.StartScan(event)
	event.Subcribe(EVENT_SCANNER, scanner.EventChannel)

	// hook os signal
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt)
	signal.Notify(osSignal, os.Kill)
	go func() {
		for sig := range osSignal {
			fmt.Println(sig.String())
			event.SendMessage(EVENT_ALL, EVENT_MAIN_TO_EXIT)
		}
	}()

	// manage event handle
	for msg := range mainEvent.Pipe {
		if msg.Status == EVENT_MAIN_TO_EXIT {

			done := event.Stop()

			// wait for end signal for event handler
		wait_loop:
			for {
				select {
				case <-done:
					break wait_loop
				case <-time.After(time.Second * 5):
					fmt.Println("timeout 5 second")
					break wait_loop
				}
			}
			scanner.StopScan()
			scanner.Close()
			fmt.Println("exit main application")
		}
	}
}

func GetVidPidFromString(str string) (vid, pid int) {
	// parse XXX:XXX to device vendorId and productId
	// split AAA:BBB to [AAA BBB]
	substringDevice := strings.Split(str, ":")

	// If length of substring is not 2, print error
	if len(substringDevice) != 2 {
		fmt.Fprintf(os.Stderr, "Error: %s is not valid format\n", str)
		flag.Usage()
		return
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

func RedirectOutput() {
	//TODO: redirect stdout to file
	//http://www.antonlindstrom.com/2014/11/17/capture-stdout-in-golang.html
}
