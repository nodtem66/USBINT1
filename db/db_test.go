package db

import (
	"fmt"
	. "github.com/nodtem66/usbint1/event"
	"testing"
	"time"
)

const (
	TEST_HOST = "10.99.3.91"
	TEST_PORT = 8086
)

func TestInflux_NewInflux(t *testing.T) {
	i := NewInflux()
	if i.Host != "localhost" || i.Port != 8086 || i.Username != "root" || i.Password != "root" {
		t.Fail()
	}

}

func TestInflux_NewInfluxWithHostPort(t *testing.T) {
	i := NewInfluxWithHostPort("a", 100)

	if i.Host != "a" || i.Port != 100 || i.Username != "root" || i.Password != "root" {
		t.Fail()
	}
}

func TestInflux_SetUserPassword(t *testing.T) {
	i := NewInflux()
	i.setUserPassword("admin", "pass")
	if i.Username != "admin" || i.Password != "pass" {
		t.Fail()
	}
}

func TestInflux_Connect(t *testing.T) {
	i := NewInfluxWithHostPort(TEST_HOST, TEST_PORT)
	i.Connect()
}

func TestInflux_QueryDatabaseName(t *testing.T) {
	i := NewInfluxWithHostPort(TEST_HOST, TEST_PORT)
	i.setUserPassword("dev", "dev")
	err := i.Connect()
	if err != nil {
		t.Fatal(err)
	}
	results, err := i.Query("SHOW DATABASES", "")
	if err != nil || results.Err != nil {
		t.Fatal(err)
	}
	for _, result := range results.Results {
		for _, row := range result.Series {
			t.Logf("column(%s) values(%s)", row.Columns[0], row.Values[0][0])
		}
	}
}

func TestInflux_StartStop(t *testing.T) {

	event := NewEventHandler()
	event.Start()

	i := NewInfluxWithHostPort(TEST_HOST, 8085)
	i.setUserPassword("dev", "dev")
	i.Start(event)
	event.SendMessage(EVENT_ALL, EVENT_MAIN_TO_EXIT)
	done := event.Stop()
	fmt.Printf("event.Stop()\n")
	<-done
}

func TestInflux_Read(t *testing.T) {
	i := NewInfluxWithHostPort(TEST_HOST, TEST_PORT)
	i.setUserPassword("dev", "dev")
	err := i.Connect()
	if err != nil {
		t.Fatal(err)
	}
	results, err := i.Query("SELECT * FROM test", "dev")
	if err != nil || results.Err != nil {
		t.Fatal(err)
	}
	if len(results.Results) == 0 {
		t.Fatal("unable to fetch data from `dev'")
	}
}

func TestInflux_Write(t *testing.T) {
	i := NewInfluxWithHostPort(TEST_HOST, TEST_PORT)
	i.setUserPassword("dev", "dev")
	err := i.Connect()
	if err != nil {
		t.Fatal()
	}
	event := NewEventHandler()
	event.Start()
	i.Start(event)

	data := &InfluxData{
		Timestamp:  time.Now(),
		PatientId:  "N1001",
		SignalType: "test",
		Resolution: 1,
		Reference:  1,
		Unit:       "",
		Data: []InfluxDataMap{
			InfluxDataMap{"a": 1},
			InfluxDataMap{"b": 2},
			InfluxDataMap{"c1": 3, "c2": 4},
		},
	}
	i.Send(data)

	event.SendMessage(EVENT_ALL, EVENT_MAIN_TO_EXIT)
	done := event.Stop()
	<-done
}
