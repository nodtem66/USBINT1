package sync

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/nodtem66/usbint1/db"
	"log"
	"strings"
	"sync"
	"time"
)

type MariaDBHandle struct {
	Connection *sql.DB
	DSN        string
	ScanRate   time.Duration
	Quit       chan struct{}
	done       chan struct{}
}

// new mysql handle from DSN and patient id
func NewMariaDBHandle(dsn string) *MariaDBHandle {
	maria := &MariaDBHandle{
		DSN:      dsn,
		Quit:     make(chan struct{}),
		done:     make(chan struct{}),
		ScanRate: time.Second,
	}
	return maria
}

// connect to external mysql server
func (m *MariaDBHandle) Connect() (err error) {
	if m.Connection, err = sql.Open("mysql", m.DSN); err != nil {
		return err
	}
	err = m.CreateTagTable()
	return
}

// create tag table in external server
func (m *MariaDBHandle) CreateTagTable() (err error) {
	createTableSql := `CREATE TABLE IF NOT EXISTS tag (
	id INTEGER NOT NULL AUTO_INCREMENT,
	patient_id VARCHAR(100) NOT NULL,
	mnt VARCHAR(50) NOT NULL,
	unit VARCHAR(50),
	resolution INT,
	ref_min DOUBLE,
	ref_max DOUBLE,
	sampling_rate INTEGER,
	descriptor TEXT NOT NULL,
	num_channel INTEGER,
	active INTEGER DEFAULT 0,
	PRIMARY KEY (id));`
	_, err = m.Connection.Exec(createTableSql)
	return
}

// create new measurement table in external server
func (m *MariaDBHandle) CreateMeasurementTable(tag DataTag, newId int64) (err error) {
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s_%s_%d ( time BIGINT NOT NULL, `,
		tag.PatientId, tag.Measurement, newId)

	// create column table from descriptor
	var i int
	for i = 0; i < tag.NumChannel-1; i++ {
		sql += strings.ToLower(tag.Descriptor[i]) + " INTEGER NOT NULL,"
	}
	sql += strings.ToLower(tag.Descriptor[i]) + " INTEGER NOT NULL, PRIMARY KEY(time));"

	// execute the sql statement
	_, err = m.Connection.Exec(sql)
	return
}

func (m *MariaDBHandle) InsertTag(tag DataTag) (newId int64, err error) {
	insertSql := `INSERT INTO tag (patient_id, mnt, unit, resolution, ref_min,
	ref_max, sampling_rate, descriptor, num_channel) 
	VALUES (?,?,?,?,?,?,?,?,?);`
	var desc []byte
	if desc, err = json.Marshal(tag.Descriptor); err != nil {
		return
	}
	var result sql.Result
	if result, err = m.Connection.Exec(insertSql, tag.PatientId, tag.Measurement, tag.Unit,
		tag.Resolution, tag.ReferenceMin, tag.ReferenceMax, tag.SamplingRate,
		string(desc), tag.NumChannel); err != nil {
		return
	}
	newId, err = result.LastInsertId()
	return
}

// prepare statement depended on number of column
// e.g. for 4 channels
// `INSERT INTO p1_ecg_1 VALUES (?,?,?,?,?);`
func (m *MariaDBHandle) PrepareStmt(tag DataTag, newId int64) (stmt *sql.Stmt, err error) {

	sql := fmt.Sprintf("INSERT INTO %s_%s_%d VALUES (?",
		tag.PatientId, tag.Measurement, newId)
	for i := 0; i < tag.NumChannel; i++ {
		sql += ",?"
	}
	sql += ");"
	stmt, err = m.Connection.Prepare(sql)
	return
}

// start worker for sychronization
func (m *MariaDBHandle) StartSyncWithPatientId(patient string) {
	ticker := time.Tick(m.ScanRate)
	go func() {
		for {
			select {
			case <-ticker:
				m.runSync(patient)
			case <-m.done:
				m.Quit <- struct{}{}
				return
			}
		}
	}()
}

// stop all worker
func (m *MariaDBHandle) Stop() {
	close(m.done)
	<-m.Quit
}

func (m *MariaDBHandle) Close() {
	m.Connection.Close()
}

func (m *MariaDBHandle) runSync(patientId string) {
	// open option for sqlite db
	dsn := fmt.Sprintf("file:%s.db?mode=rw&_busy_timeout=5000", patientId)
	// open sqlite connection
	var err error
	var sqlite *sql.DB
	var rows *sql.Rows
	if sqlite, err = sql.Open("sqlite3", dsn); err != nil {
		log.Fatal(err)
	}
	defer sqlite.Close()
	// query for measurement unit with patient id
	if rows, err = sqlite.Query(`SELECT id,mnt,unit,resolution,ref_min,ref_max,sampling_rate,descriptor,num_channel FROM tag;`); err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// store measurements from tag table into DataTag
	tags := make([]DataTag, 0)
	for rows.Next() {
		tag := DataTag{}
		var sampling int64
		var desc []byte
		if err = rows.Scan(&tag.IdTag, &tag.Measurement, &tag.Unit, &tag.Resolution, &tag.ReferenceMin, &tag.ReferenceMax, &sampling, &desc, &tag.NumChannel); err != nil {
			log.Fatal(err)
		}
		tag.PatientId = patientId
		tag.SamplingRate = time.Duration(sampling)
		json.Unmarshal(desc, &tag.Descriptor)
		tags = append(tags, tag)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	// start worker for each measurement
	var wait sync.WaitGroup
	for i := 0; i < len(tags); i++ {
		tag := tags[i]
		wait.Add(1)
		go func() {
			defer wait.Done()

			var err error
			var sqlite *sql.DB
			var rows *sql.Rows
			var newId int64
			// start worker
			log.Printf("start worker for %s_%d\n", tag.Measurement, tag.IdTag)
			// create insert tag into external server
			if newId, err = m.InsertTag(tag); err != nil {
				log.Fatal(err)
			}

			// create measurement table in external server
			if err = m.CreateMeasurementTable(tag, newId); err != nil {
				log.Fatal(err)
			}
			// prepare statement for execute
			var myStmt *sql.Stmt
			if myStmt, err = m.PrepareStmt(tag, newId); err != nil {
				log.Fatal(err)
			}
			defer myStmt.Close()
			// create sqlite connection
			if sqlite, err = sql.Open("sqlite3", dsn); err != nil {
				log.Fatal(err)
			}
			defer sqlite.Close()
			// select non-sync records order by time from low to high (ASC)
			columns := strings.Join(tag.Descriptor, ",")
			querySql := fmt.Sprintf(`SELECT time,%s FROM %s_%d WHERE sync = 0 order by time ASC;`, columns, tag.Measurement, tag.IdTag)
			if rows, err = sqlite.Query(querySql); err != nil {
				log.Fatal(err)
			}
			if _, err = m.Connection.Exec(`BEGIN;`); err != nil {
				log.Fatal(err)
			}
			for rows.Next() {
				data := make([]int64, tag.NumChannel+1)
				dataAddr := []interface{}{}
				for i := 0; i < tag.NumChannel+1; i++ {
					dataAddr = append(dataAddr, &data[i])
				}
				if err = rows.Scan(dataAddr...); err != nil {
					log.Print(err)
				}
				dataInterface := []interface{}{}
				for i := 0; i < tag.NumChannel+1; i++ {
					dataInterface = append(dataInterface, data[i])
				}
				// insert into mariadb
				myStmt.Exec(dataInterface...)
			}
			if _, err = m.Connection.Exec(`COMMIT;`); err != nil {
				log.Fatal(err)
			}
			// update record increase sync
			if _, err = sqlite.Exec(fmt.Sprintf(`UPDATE %s_%d SET sync=1 WHERE sync=0;`, tag.Measurement, tag.IdTag)); err != nil {
				log.Fatal(err)
			}
		}()
	}
	wait.Wait()
}
