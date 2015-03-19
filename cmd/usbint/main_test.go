package main

import (
	"fmt"
	. "github.com/nodtem66/usbint1"
	. "github.com/nodtem66/usbint1/event"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestMain_GetVidPidFromString(t *testing.T) {
	vid, pid := GetVidPidFromString("10C4:8846")
	if vid != 0x10C4 || pid != 0x8846 {
		t.Fatalf("error parse value 10C4:8846")
	}
	vid, pid = GetVidPidFromString("00C4:8846")
	if vid != 0x00C4 || pid != 0x8846 {
		t.Fatalf("error parse value 00C4:8846")
	}
	vid, pid = GetVidPidFromString("10C4:0046")
	if vid != 0x10C4 || pid != 0x0046 {
		t.Fatalf("error parse value 10C4:0046")
	}
	vid, pid = GetVidPidFromString("0004:0006")
	if vid != 0x0004 || pid != 0x0006 {
		t.Fatalf("error parse value 0004:0006")
	}
	vid, pid = GetVidPidFromString("14:06")
	if vid != 0x14 || pid != 0x6 {
		t.Fatalf("error parse value 14:06")
	}
	vid, pid = GetVidPidFromString("a:b")
	if vid != 0xa || pid != 0xb {
		t.Fatalf("error parse value a:b")
	}
}

func BenchmarkMain_1sec(b *testing.B) {
	benchmark_with_sec(1, b)
}
func BenchmarkMain_2sec(b *testing.B) {
	benchmark_with_sec(2, b)
}
func BenchmarkMain_3sec(b *testing.B) {
	benchmark_with_sec(3, b)
}
func BenchmarkMain_4sec(b *testing.B) {
	benchmark_with_sec(4, b)
}
func BenchmarkMain_5sec(b *testing.B) {
	benchmark_with_sec(5, b)
}
func BenchmarkMain_6sec(b *testing.B) {
	benchmark_with_sec(6, b)
}
func BenchmarkMain_7sec(b *testing.B) {
	benchmark_with_sec(7, b)
}
func BenchmarkMain_8sec(b *testing.B) {
	benchmark_with_sec(8, b)
}
func BenchmarkMain_9sec(b *testing.B) {
	benchmark_with_sec(9, b)
}
func BenchmarkMain_10sec(b *testing.B) {
	benchmark_with_sec(10, b)
}
func BenchmarkMain_13sec(b *testing.B) {
	benchmark_with_sec(13, b)
}
func BenchmarkMain_15sec(b *testing.B) {
	benchmark_with_sec(15, b)
}
func BenchmarkMain_20sec(b *testing.B) {
	benchmark_with_sec(20, b)
}

func benchmark_with_sec(timeout int, b *testing.B) {
	fmt.Printf("\n")
	var meanc, meant float32 = 0, 0
	for i := 0; i < 10; i++ {
		fmt.Printf("[%d] ", i)
		r := run_main_within_time_sec(timeout, b).([]interface{})
		for _, v := range r {
			if reflect.ValueOf(v).Kind() == reflect.String {
				vv := (string)(v.(string))
				switch vv[0] {
				case 'c':
					c, _ := strconv.Atoi(vv[1:])
					fmt.Printf(" c:%d ", c)
					//meanc = meanc*float32(i)/float32(i+1) + float32(c)/float32(i+1)
				case 't':
					t, _ := strconv.Atoi(vv[1:])
					fmt.Printf(" t:%d ", t)
					//meant = meant*float32(i)/float32(i+1) + float32(t)/float32(i+1)
				}
			}
		}
		fmt.Printf("\n")
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Printf("\nmean total_counter = %.3f\nmean total_time = %.3f\n",
		meanc, meant)
}
func run_main_within_time_sec(timeout int, b *testing.B) interface{} {

	vid, pid := GetVidPidFromString("10C4:8846")

	// create channel for exit main program
	mainEvent := NewEventSubcriptor()
	event := NewEventHandler()
	event.Start()
	event.Subcribe(EVENT_MAIN, mainEvent)

	// start scanner
	scanner := NewScanner(vid, pid)
	event.Subcribe(EVENT_SCANNER, scanner.EventChannel)

	scanner.StartScan(event)

	// setting timer to stop program
	time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		event.SendMessage(EVENT_ALL, EVENT_MAIN_TO_EXIT)
	})

	// manage event handle
	var returnValue []interface{}
	for msg := range mainEvent.Pipe {
		if msg.Status == EVENT_MAIN_TO_EXIT {

			done := event.Stop()

			// wait for end signal for event handler
		wait_loop:
			for {
				select {
				case returnValue = <-done:
					break wait_loop
				case <-time.After(time.Second * 5):
					fmt.Printf("timeout 5 second\n")
					break wait_loop
				}
			}
			scanner.StopScan()
			scanner.Close()
		}
	}
	return returnValue
}
