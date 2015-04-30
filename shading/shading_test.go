package shading

import (
	. "github.com/nodtem66/usbint1/db"
	"os"
	"testing"
	"time"
)

func TestShading_Operate(t *testing.T) {
	var err error

	sqlite := NewSqliteHandle()
	sqlite.PatientId = "TEST"
	if err = sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	if err := sqlite.EnableMeasurement([]string{"test", "test_2"}); err != nil {
		t.Fatal(err)
	}
	shade := NewHandler("./")
	shade.MinimumSync = 0
	sqlite.IdFirmware = 1
	sqlite.NumTask = 1
	sqlite.Start()
	for i := 0; i < 1000; i++ {
		sqlite.Pipe <- []int64{int64(i), 0, 1}
	}
	sqlite.Stop()
	var page_size, page_count int64
	if err = sqlite.Connection.QueryRow("PRAGMA page_size;").Scan(&page_size); err != nil {
		t.Fatal(err)
	}
	if err = sqlite.Connection.QueryRow("PRAGMA page_count;").Scan(&page_count); err != nil {
		t.Fatal(err)
	}
	old_size := page_size * page_count
	shade.Operate()
	if err = sqlite.Connection.QueryRow("PRAGMA page_size;").Scan(&page_size); err != nil {
		t.Fatal(err)
	}
	if err = sqlite.Connection.QueryRow("PRAGMA page_count;").Scan(&page_count); err != nil {
		t.Fatal(err)
	}
	new_size := page_size * page_count
	if old_size > new_size {
		t.Logf("total diffent size %d bytes", old_size-new_size)
	} else {
		t.Fatal("cannot vacuum database")
	}
}

func TestShading_Ticker(t *testing.T) {
	var err error

	sqlite := NewSqliteHandle()
	sqlite.PatientId = "TEST"
	if err = sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	//defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	if err := sqlite.EnableMeasurement([]string{"test", "test_2"}); err != nil {
		t.Fatal(err)
	}

	sqlite.IdFirmware = 1
	sqlite.NumTask = 1
	sqlite.Start()
	go func() {
		for i := 0; ; i++ {
			sqlite.Pipe <- []int64{int64(i), 0, 1}
		}
	}()
	shade := NewHandler("./")
	shade.Interval = time.Millisecond * 500
	shade.MinimumSync = 0
	shade.Start()
	time.Sleep(time.Millisecond * 5010)
	shade.Stop()
	sqlite.Stop()
	var page_size, page_count int64
	if err = sqlite.Connection.QueryRow("PRAGMA page_size;").Scan(&page_size); err != nil {
		t.Fatal(err)
	}
	if err = sqlite.Connection.QueryRow("PRAGMA page_count;").Scan(&page_count); err != nil {
		t.Fatal(err)
	}
	t.Logf("total size %d bytes", page_size*page_count)
}
