/*  Sqlite database for buffering the USB stream data
 *  Policy 1
 *    1. One Device can have multiple patientID
 *    2. One patienID can have a one sqlite db file
 *    3. One patient db file can have multiple Tag
 *    4. One Tag can have only one measurement
 *    5. One measurement can have multiple channels
 *    6. One channel can have multiple timestamp
 *    7. One timestamp can only have one value
 *
 *  Policy 2
 *    1. New Tag ID IF NO MATCH TagData AND Descriptor
 *    2. Measurement Table name := Measurement_TagID
 *
 */
/*   TAG TABLE |<-------------------TAGData ------------------------------->|
 *   --------------------------------------------------------------------------------------------------------------------
 *   | ID      | MNT   | UNIT | RESOLUTION | REFMAX | REFMIN | SamplingRate | ACTIVE  | Descriptor                      |
 *   --------------------------------------------------------------------------------------------------------------------
 *   | PRI_KEY | TEXT  | TEXT | INTEGER    |  REAL  | REAL   | INTEGER(nsec)| INTEGER | TEXT                            |
 *   --------------------------------------------------------------------------------------------------------------------
 *   e.g.
 *   --------------------------------------------------------------------------------------------------------------------
 *   | 1       | ECG1  |  mV  | 2048       |  1     |  -1    | 1000         | 0       | {"1": "LEAD_I", "2": "LEAD_II"} |
 *   --------------------------------------------------------------------------------------------------------------------
 *   | 2       | ECG2  |   V  | 1024       |  5     |   0    | 2000         | 1       | {"1": "LEAD_VI", "4": "LEAD_II"}|
 *   --------------------------------------------------------------------------------------------------------------------
 *   | 3       | SPO2  |  %   | 1024       |  0     |  100   | 3000         | 0       | {"5": "LEAD_I", "10": "LEAD_II"}|
 *   --------------------------------------------------------------------------------------------------------------------
 *  MNT : Measurement
 *  CHN : Channel
 *  TABLE_NAMES : ECG1_1, ECG2_2, SPO2_3
 */

/*   Structure MEASUREMENT Table
 *
 *   Table Name (see Policy 2.2) e.g. ECG_1
 *   --------------------------------------------
 *   | TIME    | CHANNEL_ID | VALUE   | TAG_ID  |
 *   --------------------------------------------
 *   | PRI_KEY | INTEGER    | INTEGER | INTEGER |
 *   --------------------------------------------
 *   e.g.
 *   --------------------------------------------
 *   | 109880980980  | 1    |  10908  | 1       |
 *   --------------------------------------------
 *   | 1988-09-0909  | 2    |  78909  | 1       |
 *   --------------------------------------------
 */
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/nodtem66/usbint1/event"
	"os"
	"strings"
	"sync"
	"time"
)

type SqliteData []int64
type SqliteHandle struct {
	Connection   *sql.DB
	EventChannel *EventSubscriptor
	IdTag        int64
	Pipe         chan []int64
	Quit         chan bool
	WaitQuit     sync.WaitGroup
	NumTask      int
	TimeStamp    time.Time
	DataTag
}

func NewSqliteHandle() *SqliteHandle {
	sqlite := &SqliteHandle{
		DataTag: DataTag{
			PatientId:    "none",
			Measurement:  "general",
			ReferenceMax: 1,
			SamplingRate: time.Millisecond,
		},
		EventChannel: NewEventSubcriptor(),
		Pipe:         make(chan []int64, LENGTH_QUEUE),
		Quit:         make(chan bool),
		NumTask:      1,
	}
	return sqlite
}

func (s *SqliteHandle) Connect() (err error) {

	// set the database name according to patientId
	database_name := "./" + s.PatientId + ".db"
	s.Connection, err = sql.Open("sqlite3", database_name)
	if err != nil {
		return err
	}

	if err = s.CreateTagTable(); err != nil {
		return err
	}

	return nil
}
func (s *SqliteHandle) ConnectNew() error {

	// set the database name according to patientId
	database_name := "./" + s.PatientId + ".db"

	// check database file exitsts
	if _, err := os.Stat(database_name); err == nil {

		// Bugs: text file busy
		// https://github.com/jnwhiteh/golang/blob/master/src/cmd/go/build.go
		nbusy := 0
		var removeErr error
		// trying to remove file three time
		// If `text file busy` error happens, sleep a little
		// and try again.  We let this happen three times, with increasing
		// sleep lengths: 100+200+400 ms = 0.7 seconds.
		for {
			removeErr = os.Remove(database_name)
			if removeErr != nil && strings.Contains(removeErr.Error(), "text file busy") && nbusy < 3 {
				time.Sleep(100 * time.Millisecond << uint(nbusy))
				nbusy++
				continue
			} else {
				break
			}
		}
		if removeErr != nil {
			return removeErr
		}
	}

	return s.Connect()
}

func (s *SqliteHandle) CreateTagTable() error {
	// For optmize use PRAGMA journal_mod=WAL
	if _, err := s.Connection.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return err
	}
	// For optmize use PRAGMA synchronous=NORMAL
	/* Create table tag
	   TAG TABLE |<-------------------TAGData ------------------------------->|
	   --------------------------------------------------------------------------------------------------------------------
	   | ID      | MNT   | UNIT | RESOLUTION | REFMAX | REFMIN | SamplingRate | ACTIVE  | Descriptor                      |
	   --------------------------------------------------------------------------------------------------------------------
	   | PRI_KEY | TEXT  | TEXT | INTEGER    |  REAL  | REAL   | INTEGER(nsec)| INTEGER | TEXT                            |
	   --------------------------------------------------------------------------------------------------------------------
	   e.g.
	   --------------------------------------------------------------------------------------------------------------------
	   | 1       | ECG1  |  mV  | 2048       |  1     |  -1    | 1000         | 0       | {"1": "LEAD_I", "2": "LEAD_II"} |
	   --------------------------------------------------------------------------------------------------------------------
	   | 2       | ECG2  |   V  | 1024       |  5     |   0    | 2000         | 1       | {"1": "LEAD_VI", "4": "LEAD_II"}|
	   --------------------------------------------------------------------------------------------------------------------
	   | 3       | SPO2  |  %   | 1024       |  0     |  100   | 3000         | 0       | {"5": "LEAD_I", "10": "LEAD_II"}|
	   --------------------------------------------------------------------------------------------------------------------
	   MNT : Measurement
	   CHN : Channel
	*/
	tagTableStmt := `CREATE TABLE IF NOT EXISTS tag (
	id INTEGER NOT NULL PRIMARY KEY,
	mnt TEXT NOT NULL,
	unit TEXT,
	resolution INTEGER,
	ref_min INTEGER,
	ref_max INTEGER,
	sampling_rate INTEGER,
	descriptor TEXT NOT NULL,
	active INTEGER DEFAULT 0);`

	if _, err := s.Connection.Exec(tagTableStmt); err != nil {
		return err
	}

	return nil
}

/* EnableMeasurement enable the current measurement with tagId and descriptor
 * Param: DescriptionType
 * Example:
 * EnableMeasurement([]string{"LEAD_I", "LEAD_II"})
 */
func (s *SqliteHandle) EnableMeasurement(desc []string) error {

	if len(desc) == 0 {
		fmt.Errorf("Empty DescriptorType")
	}
	// convert to json
	jsonDdesc, _ := json.Marshal(desc)

	queryMeasurementStmt := `SELECT id FROM tag WHERE 
	mnt = ? AND unit = ? AND resolution = ? AND ref_min = ? AND 
	ref_max = ? AND sampling_rate = ? AND descriptor= ?;`

	insertMeasurementStmt := `INSERT INTO tag 
	(mnt, unit, resolution, ref_min, ref_max, sampling_rate, descriptor, active)
	VALUES (?,?,?,?,?,?,?,?);`

	updateMeasurementStmt := `UPDATE tag SET active = 1 WHERE id = ?;`

	// Prepare SQL statement for Search matched TagID
	p, err := s.Connection.Prepare(queryMeasurementStmt)
	if err != nil {
		return err
	}

	// Find the previous measurement in table tag
	err = p.QueryRow(s.Measurement,
		s.Unit,
		s.Resolution,
		s.ReferenceMin,
		s.ReferenceMax,
		s.SamplingRate,
		string(jsonDdesc)).Scan(&s.IdTag)
	switch {
	case err == sql.ErrNoRows:
		// Insert new measurement in table tag and enable it
		p, err := s.Connection.Prepare(insertMeasurementStmt)
		if err != nil {
			return err
		}
		result, err := p.Exec(s.Measurement, s.Unit, s.Resolution, s.ReferenceMin, s.ReferenceMax, s.SamplingRate, string(jsonDdesc), 1)
		if err != nil {
			return err
		}
		if id, err := result.LastInsertId(); err == nil {
			s.IdTag = id
		} else {
			return err
		}
		p.Close()
	case err != nil:
		return nil
	default:
		// Enables measurement in table tag
		p, err := s.Connection.Prepare(updateMeasurementStmt)
		if err != nil {
			return err
		}
		if _, err := p.Exec(s.IdTag); err != nil {
			return err
		}
		p.Close()
	}
	p.Close()

	// Create MEASUREMENT Table
	/*   Structure MEASUREMENT Table
	 *   --------------------------------------------
	 *   | TIME    | CHANNEL_ID | VALUE   | TAG_ID  |
	 *   --------------------------------------------
	 *   | PRI_KEY | INTEGER    | INTEGER | INTEGER |
	 *   --------------------------------------------
	 *   e.g.
	 *   --------------------------------------------
	 *   | 109880980980  | 1    |  10908  | 1       |
	 *   --------------------------------------------
	 *   | 1988-09-0909  | 2    |  78909  | 1       |
	 *   --------------------------------------------
	 */
	measurementTableStmt := `CREATE TABLE IF NOT EXISTS %s_%d (
	time INTEGER NOT NULL,
	channel_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL,
	value INTEGER NOT NULL,
	PRIMARY KEY (time, channel_id, tag_id));`
	_, err = s.Connection.Exec(fmt.Sprintf(measurementTableStmt, s.Measurement, s.IdTag))
	if err != nil {
		return err
	}
	return nil
}
func (s *SqliteHandle) DisableMeasurement() error {
	updateMeasurementStmt := `UPDATE tag SET active = 0 WHERE id = ?;`
	p, err := s.Connection.Prepare(updateMeasurementStmt)
	if err != nil {
		return err
	}
	if _, err := p.Exec(s.IdTag); err != nil {
		return err
	}
	p.Close()
	return nil
}

// Must be start after EnableMeasurement
func (s *SqliteHandle) Start() {

	// create SQL insertion
	insertStmt := `INSERT INTO %s_%d (time, channel_id, tag_id, value) VALUES (?,?,?,?);`
	insertStmt = fmt.Sprintf(insertStmt, s.Measurement, s.IdTag)

	// init worker
	for i := 0; i < s.NumTask; i++ {
		s.WaitQuit.Add(1)
		// main routine
		go func() {
			//s.Quit <- true
			defer s.WaitQuit.Done()
			// main loop
			for data := range s.Pipe {

				// Transaction
				tx, err := s.Connection.Begin()
				if err != nil {
					fmt.Println("Err TX Begin(): ", err)
				}
				stmt, err := tx.Prepare(insertStmt)
				if err != nil {
					fmt.Println("Err TX Prepare(): ", err)
				}

				// Reset timeout
				/*
					isTimeout := time.Now().Sub(s.TimeStamp) > s.SamplingRate*5
					if isTimeout {
						s.TimeStamp = time.Now()
					}*/
				timestamp := data[0]
				data = data[1:]
				for i, d := range data {
					if _, err := stmt.Exec(timestamp, i, s.IdTag, d); err != nil {
						fmt.Println("Err TX Exec: ", err)
					}
				}
				//s.TimeStamp = s.TimeStamp.Add(s.SamplingRate)
				if err := tx.Commit(); err != nil {
					fmt.Println("Err TX Commit(): ", err)
				}
			}
		}()
	}

	go func() {
		s.WaitQuit.Wait()
		s.Quit <- true
	}()

	// event routine
	go func() {
		for msg := range s.EventChannel.Pipe {
			if msg.Name == EVENT_DATABASE && msg.Status == EVENT_DATABASE_TO_EXIT {
				s.Stop()
				s.EventChannel.Done <- struct{}{}
			}
		}
	}()
}

func (s *SqliteHandle) Stop() {
	close(s.Pipe)
	<-s.Quit
}

func (s *SqliteHandle) Send(data []int64) {
	s.Pipe <- data
}

func (s *SqliteHandle) Close() {
	if err := s.DisableMeasurement(); err != nil {
		fmt.Printf("%s", err)
	}
	s.Connection.Close()
}
