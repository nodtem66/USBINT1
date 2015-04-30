package db

import (
	"os"
	"testing"
)

func TestSqlite_Connect(t *testing.T) {
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

	var b int
	err := sqlite.Connection.QueryRow(`SELECT active FROM tag WHERE mnt='` + sqlite.Measurement + `';`).Scan(&b)
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

	err = sqlite.Connection.QueryRow(`SELECT active FROM tag WHERE mnt='` + sqlite.Measurement + `';`).Scan(&b)
	if err != nil {
		t.Fatal(err)
	}
	if b != 0 {
		t.Fatalf("Unexceptional Active Tag for `%s` (%d)", sqlite.Measurement, b)
	}
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
	if _, err := sqlite.Connection.Exec(createSqlStmt); err != nil {
		t.Fatal(err)
	}
	// insert the data
	stmt, err := sqlite.Connection.Prepare(insertSqlStmt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stmt.Exec("hello world"); err != nil {
		t.Fatal(err)
	}
	stmt.Close()
	// query the data
	rows, err := sqlite.Connection.Query(querySqlStmt)
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
func TestSqlite_Pragma(t *testing.T) {
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

	rows, err := sqlite.Connection.Query(`PRAGMA synchronou;`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var intSync int
		if err := rows.Scan(&intSync); err != nil {
			t.Fatal(err)
		}
		if intSync != 1 {
			t.Fatal("PRAGMA = ", intSync)
		}

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

	rows, err := sqlite.Connection.Query(`SELECT name FROM sqlite_master WHERE type='table';`)
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

	p, err := sqlite.Connection.Prepare(`SELECT descriptor FROM tag WHERE id = ?`)
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
	if err := sqlite.Connect(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()
	if err := sqlite.EnableMeasurement([]string{"id", "temperature"}); err != nil {
		t.Fatal(err)
	}
	sqlite.Start()
	sqlite.Stop()
}

func TestSqlite_SendViaPipe(t *testing.T) {
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

	sqlite.Pipe <- []int64{1, 1}
	sqlite.Pipe <- []int64{2, 1}
	sqlite.Pipe <- []int64{3, 2}
	sqlite.Pipe <- []int64{4, 3}
	sqlite.Pipe <- []int64{5, 4}
	sqlite.Pipe <- []int64{1024, 11}

	sqlite.Stop()
	rows, err := sqlite.Connection.Query(`SELECT time, test value FROM general_1;`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var time, test_value int64
		if err := rows.Scan(&time, &test_value); err != nil {
			t.Fatal(err)
		}
		t.Logf("[%d] %d\n", time, test_value)
	}
}

func TestSqlite_SendTask(t *testing.T) {
	sqlite := NewSqliteHandle()
	sqlite.PatientId = "T001"
	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	if err := sqlite.EnableMeasurement([]string{"test", "test2", "test3"}); err != nil {
		t.Fatal(err)
	}
	t.Log("NumTask: ", sqlite.NumTask)
	sqlite.Start()
	var i int64
	for i = 0; i < 1000; i++ {
		sqlite.Pipe <- []int64{i, 1, 0, 0}
	}

	sqlite.Stop()
	var total int
	if err := sqlite.Connection.QueryRow(`SELECT count(*) FROM general_1;`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	t.Log("Total insert: ", total)
}

func TestSqlite_SendOneTask(t *testing.T) {
	sqlite := NewSqliteHandle()
	sqlite.PatientId = "T001"
	if err := sqlite.ConnectNew(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sqlite.PatientId + ".db")
	defer sqlite.Close()

	if err := sqlite.EnableMeasurement([]string{"test", "test2", "test3"}); err != nil {
		t.Fatal(err)
	}
	sqlite.NumTask = 1
	sqlite.Start()
	var i int64
	for i = 0; i < 3000; i++ {
		sqlite.Pipe <- []int64{i, 1, 0, 0}
	}

	sqlite.Stop()
	var total int
	if err := sqlite.Connection.QueryRow(`SELECT count(*) FROM general_1;`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	t.Log("Total insert: ", total)
}
