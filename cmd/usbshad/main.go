package main

import (
	"github.com/BurntSushi/toml"
	. "github.com/nodtem66/usbint1/config"
	. "github.com/nodtem66/usbint1/shading"
	"log"
	"os"
	"os/signal"
)

// Define Build LDFLAGS variable
var Version string
var Commit string

func main() {

	// read toml config file
	conf := TomlConfig{
		DB:    Database{Path: "./"},
		Shade: Shading{MinimumSync: 1},
	}
	if len(os.Args) > 1 {
		if _, err := toml.DecodeFile(os.Args[1], &conf); err != nil {
			log.Fatal(err)
		}
	} else {
		if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
			log.Fatal(err)
		}
	}

	// redirect log to file
	if len(conf.Log.FileName) != 0 {

		// open file
		logfile, err := os.OpenFile(conf.Log.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer logfile.Close()

		// setting file to log
		log.SetOutput(logfile)
	}

	// create shading
	shade := NewHandler(conf.DB.Path)
	shade.Interval = conf.Shade.Interval.Duration
	shade.MinimumSync = conf.Shade.MinimumSync

	// hook os signal for exit program
	osSignal := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(osSignal, os.Kill)
	signal.Notify(osSignal, os.Interrupt)

	go func() {
		for _ = range osSignal {
			shade.Stop()
			done <- true
		}
	}()
	log.SetPrefix("[USB_SHADE] ")
	log.Printf("[Start Cleaning every %s] [database scaning path %s]\n",
		conf.Shade.Interval, conf.DB.Path)
	shade.Start()

	//wait for exit signal
	<-done
	log.Printf("[Stop]")
}
