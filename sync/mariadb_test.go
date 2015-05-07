package sync

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	. "github.com/nodtem66/usbint1/db"
	"os"
	"testing"
	"time"
)

// Data Source Name
// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
const DSN = "root:root@tcp(127.0.0.1:3306)/department1"

func TestMariaDB_Connect(t *testing.T) {
	var err error
	maria := NewMariaDBHandle(DSN)
	if err = maria.Connect(); err != nil {
		t.Fatal(err)
	}
	defer maria.Close()

	if results, err := maria.Connection.Query(`SHOW DATABASES;`); err != nil {
		t.Fatal(err)
	} else {
		tables := make([]string, 0)
		for results.Next() {
			var table string
			if err := results.Scan(&table); err != nil {
				t.Fatal(err)
			}
			tables = append(tables, table)
		}
		if err := results.Err(); err != nil {
			t.Fatal(err)
		}
		t.Log(tables)
	}
	// drop table tag and measurement
	if _, err = maria.Connection.Exec(`DROP TABLE IF EXISTS tag;`); err != nil {
		t.Fatal(err)
	}
}

func TestMariaDB_InsertTag(t *testing.T) {
	var err error
	maria := NewMariaDBHandle(DSN)
	if err = maria.Connect(); err != nil {
		t.Fatal(err)
	}
	defer maria.Close()
	tag := DataTag{
		PatientId:    "test",
		NumChannel:   2,
		Descriptor:   []string{"a", "b"},
		Measurement:  "test1",
		Unit:         "mB",
		Resolution:   1000,
		SamplingRate: 1000,
		ReferenceMin: 0,
		ReferenceMax: 1,
	}
	var newId int64
	if newId, err = maria.InsertTag(tag); err != nil {
		t.Fatal(err)
	}
	t.Log(newId)
	// drop table tag and measurement
	if _, err = maria.Connection.Exec(`DROP TABLE IF EXISTS tag;`); err != nil {
		t.Fatal(err)
	}
}

func TestMariaDB_CreateMeasurement(t *testing.T) {
	var err error
	maria := NewMariaDBHandle(DSN)
	if err = maria.Connect(); err != nil {
		t.Fatal(err)
	}
	defer maria.Close()
	tag := DataTag{
		PatientId:    "test",
		NumChannel:   2,
		Descriptor:   []string{"a", "b"},
		Measurement:  "test2",
		Unit:         "mB",
		Resolution:   1000,
		SamplingRate: 1000,
		ReferenceMin: 0,
		ReferenceMax: 1,
	}
	var newId int64
	if newId, err = maria.InsertTag(tag); err != nil {
		t.Fatal(err)
	}
	if err = maria.CreateMeasurementTable(tag, newId); err != nil {
		t.Fatal(err)
	}
	// drop table tag and measurement
	if _, err = maria.Connection.Exec(`DROP TABLE IF EXISTS tag, test_test2_1;`); err != nil {
		t.Fatal(err)
	}
}

func TestMariaDB_InsertMeasurement(t *testing.T) {
	var err error
	var newId int64

	maria := NewMariaDBHandle(DSN)
	if err = maria.Connect(); err != nil {
		t.Fatal(err)
	}
	defer maria.Close()
	tag := DataTag{
		PatientId:    "test",
		NumChannel:   2,
		Descriptor:   []string{"c", "b"},
		Measurement:  "test3",
		Unit:         "mB",
		Resolution:   1000,
		SamplingRate: 1000,
		ReferenceMin: 0,
		ReferenceMax: 1,
	}
	if newId, err = maria.InsertTag(tag); err != nil {
		t.Fatal(err)
	}
	if err = maria.CreateMeasurementTable(tag, newId); err != nil {
		t.Fatal(err)
	}
	// prepare statement for execute
	var stmt *sql.Stmt
	if stmt, err = maria.PrepareStmt(tag, newId); err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 10; i++ {
		if _, err = stmt.Exec([]interface{}{time.Now().UnixNano(), 0, 1}...); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err = maria.Connection.QueryRow(`SELECT COUNT(*) FROM test_test3_1;`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 10 {
		t.Log(count)
	}
	// drop table tag and measurement
	if _, err = maria.Connection.Exec(`DROP TABLE IF EXISTS tag, test_test3_1;`); err != nil {
		t.Fatal(err)
	}
}
func TestMariaDB_runSync(t *testing.T) {
	var err error

	// init sqlite connection
	sqlite := NewSqliteHandle()
	sqlite.PatientId = "TEST"
	if err = sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	// init mysql connection
	maria := NewMariaDBHandle(DSN)
	if err = maria.Connect(); err != nil {
		t.Fatal(err)
	}
	defer maria.Close()

	if err := sqlite.EnableMeasurement([]string{"test", "test_2"}); err != nil {
		t.Fatal(err)
	}

	sqlite.IdFirmware = 1
	sqlite.NumTask = 1
	sqlite.Start()
	var count int64
	go func() {
		for i := 0; ; i++ {
			sqlite.Pipe <- []int64{int64(i), 0, 1}
			count++
		}
	}()
	time.Sleep(10 * time.Millisecond)
	sqlite.Stop()
	maria.runSync("TEST", "./TEST.db")

	// drop table tag and measurement
	var count2 int64
	if err = maria.Connection.QueryRow(`SELECT count(*) FROM TEST_general_1;`).Scan(&count2); err != nil {
		t.Fatal(err)
	}
	var count3 int64
	if err = sqlite.Connection.QueryRow(`SELECT count(*) FROM general_1 WHERE sync=0;`).Scan(&count3); err != nil {
		t.Fatal(err)
	}
	t.Logf("[total read: %d write: %d left: %d]", count, count2, count3)
	if _, err = maria.Connection.Exec(`DROP TABLE IF EXISTS tag, TEST_general_1;`); err != nil {
		t.Fatal(err)
	}
}
func TestMariaDB_StartStop(t *testing.T) {
	var err error

	// init sqlite connection
	sqlite := NewSqliteHandle()
	sqlite.PatientId = "TEST"
	if err = sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	// init mysql connection
	maria := NewMariaDBHandle(DSN)
	maria.ScanRate = time.Second
	if err = maria.Connect(); err != nil {
		t.Fatal(err)
	}
	defer maria.Close()
	maria.Start()

	if err := sqlite.EnableMeasurement([]string{"test", "test_2"}); err != nil {
		t.Fatal(err)
	}

	sqlite.IdFirmware = 1
	sqlite.NumTask = 1
	sqlite.Start()
	var count int64
	go func() {
		for i := 0; ; i++ {
			sqlite.Pipe <- []int64{int64(i), 0, 1}
			count++
		}
	}()
	time.Sleep(10 * time.Millisecond)
	sqlite.Stop()
	time.Sleep(time.Second)
	maria.Stop()

	// drop table tag and measurement
	var count2 int64
	if err = maria.Connection.QueryRow(`SELECT count(*) FROM TEST_general_1;`).Scan(&count2); err != nil {
		t.Fatal(err)
	}
	var count3 int64
	if err = sqlite.Connection.QueryRow(`SELECT count(*) FROM general_1 WHERE sync=0;`).Scan(&count3); err != nil {
		t.Fatal(err)
	}
	t.Logf("[total read: %d write: %d left: %d]", count, count2, count3)
	if _, err = maria.Connection.Exec(`DROP TABLE IF EXISTS tag, TEST_general_1;`); err != nil {
		t.Fatal(err)
	}
}
