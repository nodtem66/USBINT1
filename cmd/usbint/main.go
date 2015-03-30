package main

import (
	"flag"
	"fmt"
	. "github.com/nodtem66/usbint1"
	"github.com/nodtem66/usbint1/config"
	"github.com/nodtem66/usbint1/db"
	. "github.com/nodtem66/usbint1/event"
	"os"
	"os/signal"
	"time"
)

const (
	DEFAULT_DEVICE_VID_PID string = "10C4:8846"
	DEFAULT_INFLUXDB_HOST  string = "127.0.0.1:8086"
)

var (
	deviceString = flag.String("dev", DEFAULT_DEVICE_VID_PID, "device to "+
		"listen VENDOR:PRODUCT (hex).")
	hostInfluxDBString = flag.String("influxdb", DEFAULT_INFLUXDB_HOST, "Influxdb API address "+
		"to host the streaming data.")
	patientId = flag.String("patient", "", "patient id to store as measurement unit in influxdb")
)

func main() {
	//get last execute name
	programName := config.GetProgramName(os.Args[0])

	//replace default usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", programName)
		flag.PrintDefaults()
	}
	flag.Parse()

	// check there is patientID
	if len(*patientId) == 0 {
		fmt.Fprintf(os.Stderr, "Error: patient cannot be null string\n")
		flag.Usage()
		return
	}

	// parse vendorId, productId from command option
	vid, pid, err := config.GetVidPidFromString(*deviceString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		flag.Usage()
		return
	}

	// create channel for exit main program
	mainEvent := NewEventSubcriptor()
	event := NewEventHandler()
	event.Start()
	event.Subcribe(EVENT_MAIN, mainEvent)

	// parse host:post parameters from command option
	host, port, err := config.GetHostPortFromString(*hostInfluxDBString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		flag.Usage()
		return
	}

	// print infomation about daemon
	fmt.Printf("Patiend ID: %s\nConnect to influxdb %s:%d\n", *patientId, host, port)
	fmt.Printf("Initialize USB scanner to device %04X:%04X\n", vid, pid)

	// start database interface
	// influx := db.NewInfluxWithHostPort(host, port)
	// influx.PatientId = *patientId
	// influx.Start(event)

	// start scanner
	scanner := NewScanner(vid, pid)
	defer scanner.Close()
	scanner.StartScan(event, influx)

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
			//scanner.StopScan()
			fmt.Println("exit main application")
			return
		}
	}
}

func RedirectOutput() {
	//TODO: redirect stdout to file
	//http://www.antonlindstrom.com/2014/11/17/capture-stdout-in-golang.html
}
