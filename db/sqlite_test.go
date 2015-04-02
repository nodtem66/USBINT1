package db

import (
	. "github.com/nodtem66/usbint1/event"
	"os"
	"testing"
)

func TestSqlite_Connect(t *testing.T) {
	sqlite := NewSqliteHandle()

	sqlite.PatientId = "T001"
	if err := sqlite.Connect(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	if err := sqlite.EnableMeasurement([]string{"test"}); err != nil {
		t.Fatal(err)
	}

	var b int
	err := sqlite.connection.QueryRow(`SELECT active FROM tag WHERE mnt='` + sqlite.Measurement + `';`).Scan(&b)
	if err != nil {
		t.Fatal(err)
	}
	if b != 1 {
		t.Fatalf("Unexceptional Deactive Tag for `%s` (%d)", sqlite.Measurement, b)
	}
	sqlite.Close()
	if err := sqlite.Connect(); err != nil {
		t.Fatal(err)
	}

	err = sqlite.connection.QueryRow(`SELECT active FROM tag WHERE mnt='` + sqlite.Measurement + `';`).Scan(&b)
	if err != nil {
		t.Fatal(err)
	}
	if b != 0 {
		t.Fatalf("Unexceptional Active Tag for `%s` (%d)", sqlite.Measurement, b)
	}

	sqlite.Close()

}

func TestSqlite_CreateTable(t *testing.T) {

	sqlite := NewSqliteHandle()
	sqlite.PatientId = "TEST"

	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	createSqlStmt := `CREATE TABLE IF NOT EXISTS test (id INTEGER NOT NULL PRIMARY KEY, name TEXT);`
	insertSqlStmt := `INSERT INTO test (name) VALUES (?);`
	querySqlStmt := `SELECT * FROM test LIMIT 10;`

	// create a table
	if _, err := sqlite.connection.Exec(createSqlStmt); err != nil {
		t.Fatal(err)
	}
	// insert the data
	stmt, err := sqlite.connection.Prepare(insertSqlStmt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stmt.Exec("hello world"); err != nil {
		t.Fatal(err)
	}
	stmt.Close()
	// query the data
	rows, err := sqlite.connection.Query(querySqlStmt)
	defer rows.Close()
	for rows.Next() {
		var name string
		var id int
		if err := rows.Scan(&id, &name); err != nil {
			t.Fatal(err)
		}
		t.Logf("record (%d, %s)", id, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestSqlite_NewMeasurementTable(t *testing.T) {
	sqlite := NewSqliteHandle()
	sqlite.PatientId = "TEST"
	sqlite.ReferenceMin = 0
	sqlite.ReferenceMax = 5
	sqlite.Unit = "C"
	sqlite.Resolution = 1024

	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	if err := sqlite.EnableMeasurement([]string{"test"}); err != nil {
		t.Fatal(err)
	}

	rows, err := sqlite.connection.Query(`SELECT name FROM sqlite_master WHERE type='table';`)
	if err != nil {
		t.Fatal(err)
	}
	checkTable := map[string]bool{
		"tag":       false,
		"general_1": false,
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		t.Log(name)
		if checkTable[name] == false {
			checkTable[name] = true
		} else {
			t.Fatalf("Duplicate table `%s`\n", name)
		}
	}
	for key, value := range checkTable {
		if value == false {
			t.Fatalf("table %s didn't found\n", key)
		}
	}
}

func TestSqlite_EnableMeasurement(t *testing.T) {
	sqlite := NewSqliteHandle()
	sqlite.PatientId = "TEST"
	sqlite.ReferenceMin = 0
	sqlite.ReferenceMax = 5
	sqlite.Unit = "C"
	sqlite.Resolution = 1024

	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	names := []string{
		"LEAD I", "LEAD II", "LEAD III",
	}
	sqlite.EnableMeasurement(names)

	p, err := sqlite.connection.Prepare(`SELECT descriptor FROM tag WHERE id = ?`)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := p.Query(sqlite.IdTag)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var desc string
		err := rows.Scan(&desc)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(desc)
	}
	p.Close()
}

func TestSqlite_StartStop(t *testing.T) {
	sqlite := NewSqliteHandle()
	sqlite.PatientId = "T001"
	sqlite.Start()
	sqlite.Stop()
}

func TestSqlite_StartStopWithEventManager(t *testing.T) {
	event := NewEventHandler()
	event.Start()

	sqlite := NewSqliteHandle()
	sqlite.PatientId = "T001"
	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	if err := sqlite.EnableMeasurement([]string{"test"}); err != nil {
		t.Fatal(err)
	}
	sqlite.Start()
	event.Subcribe(EVENT_DATABASE, sqlite.EventChannel)
	event.SendMessage(EVENT_DATABASE, EVENT_DATABASE_TO_EXIT)
	wait := event.Stop()
	<-wait
}

func TestSqlite_Send(t *testing.T) {
	event := NewEventHandler()
	event.Start()

	sqlite := NewSqliteHandle()
	sqlite.PatientId = "T001"
	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	if err := sqlite.EnableMeasurement([]string{"ch1", "ch2", "ch3"}); err != nil {
		t.Fatal(err)
	}
	sqlite.Start()
	event.Subcribe(EVENT_DATABASE, sqlite.EventChannel)
	sqlite.Send(SqliteData{1, 2, 3})

	event.SendMessage(EVENT_DATABASE, EVENT_DATABASE_TO_EXIT)
	wait := event.Stop()
	<-wait

	rows, err := sqlite.connection.Query(`SELECT time, channel_id, tag_id, value FROM general_1;`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var n, time, c, tid int
		if err := rows.Scan(&time, &c, &tid, &n); err != nil {
			t.Fatal(err)
		}
		t.Logf("[%d|%d:%d] %d\n", time, tid, c, n)
	}
}

func TestSqlite_SendViaPipe(t *testing.T) {
	event := NewEventHandler()
	event.Start()

	sqlite := NewSqliteHandle()
	sqlite.PatientId = "T001"
	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	if err := sqlite.EnableMeasurement([]string{"test"}); err != nil {
		t.Fatal(err)
	}
	sqlite.Start()
	event.Subcribe(EVENT_DATABASE, sqlite.EventChannel)
	event.SendMessage(EVENT_DATABASE, EVENT_DATABASE_TO_EXIT)
	wait := event.Stop()
	<-wait
}
