package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"
	. "github.com/nodtem66/usbint1/config"
	. "github.com/nodtem66/usbint1/sync"
)

// Define Build LDFLAGS variable
var Version string
var Commit string

func main() {

	// read toml config file
	conf := TomlConfig{
		DB:    Database{Path: "./"},
		Shade: Shading{MinimumSync: 1},
		Sync:  SyncInfo{Mode: "shade"},
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
	//shade := NewHandler(conf.DB.Path)
	//shade.Interval = conf.Shade.Interval.Duration
	//shade.MinimumSync = conf.Shade.MinimumSync
	var err error
	maria := NewMariaDBHandle(conf.Sync.DSN)
	if conf.Sync.Mode == "shade" {
	} else if conf.Sync.Mode == "sync" {
		if err = maria.Connect(); err != nil {
			log.Fatal(err)
		}
		defer maria.Close()
	}
	maria.Mode = conf.Sync.Mode
	maria.ScanRate = conf.Sync.Interval.Duration
	maria.ShadeTime = conf.Sync.ShadeTime.Duration
	maria.Root = conf.DB.Path

	// hook os signal for exit program
	osSignal := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(osSignal, os.Kill)
	signal.Notify(osSignal, os.Interrupt)

	go func() {
		for _ = range osSignal {
			maria.Stop()
			log.Printf("[Stop]")
			done <- true
		}
	}()
	log.SetPrefix("[USB_SYNC] ")
	log.Printf("[Start sync every %s/%s] [database scaning path %s]\n",
		maria.ScanRate, maria.ShadeTime, conf.DB.Path)
	maria.Start()

	//wait for exit signal
	<-done
}
