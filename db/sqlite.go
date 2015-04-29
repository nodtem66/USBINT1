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
 *   ----------------------------------------------------------------------------------------------------------------------------------
 *   | ID      | MNT   | UNIT | RESOLUTION | REFMAX | REFMIN | SamplingRate | ACTIVE  | Descriptor                      | Num Channel |
 *   ----------------------------------------------------------------------------------------------------------------------------------
 *   | PRI_KEY | TEXT  | TEXT | INTEGER    |  REAL  | REAL   | INTEGER(nsec)| INTEGER | TEXT                            | INTEGER     |
 *   ----------------------------------------------------------------------------------------------------------------------------------
 *   e.g.
 *   ---------------------------------------------------------------------------------------------------------------------------------
 *   | 1       | ECG1  |  mV  | 2048       |  1     |  -1    | 1000         | 0       | {"1": "LEAD_I", "2": "LEAD_II"} | 2          |
 *   ---------------------------------------------------------------------------------------------------------------------------------
 *   | 2       | ECG2  |   V  | 1024       |  5     |   0    | 2000         | 1       | {"1": "LEAD_VI", "4": "LEAD_II"}| 2          |
 *   ---------------------------------------------------------------------------------------------------------------------------------
 *   | 3       | SPO2  |  %   | 1024       |  0     |  100   | 3000         | 0       | {"5": "LEAD_I", "10": "LEAD_II"}| 2          |
 *   ---------------------------------------------------------------------------------------------------------------------------------
 *  MNT : Measurement
 *  CHN : Channel
 *  TABLE_NAMES : ECG1_1, ECG2_2, SPO2_3
 */

/*   Structure MEASUREMENT Table
 *
 *   Table Name (see Policy 2.2) e.g. ECG_1
 *   ---------------------------------------------------------------------
 *   | TIME (nsec)  | SYNC    | CHANNEL_1 | CHANNEL_2 | .... | CHANNEL_N |
 *   ---------------------------------------------------------------------
 *   | PRI_KEY      | INTEGER | INTEGER   | INTEGER   | .... |  INTEGER  |
 *   ---------------------------------------------------------------------
 *   e.g.
 *   ----------------------------------------------------------------------
 *   | 109880980980 | 0       | 10908     | 1         | ...... | ..       |
 *   ----------------------------------------------------------------------
 *   | 1988-09-0909 | 0       | 78909     | 1         | ...... | ..       |
 *   ----------------------------------------------------------------------
 */

/* Example usage:
1. sqlite := NewSqliteHandle()
2. sqlite.Connect()
   or sqlite.ConnectNew()
3. sqlite.EnableMeasurement()
4. sqlite.IdFirmware = ...
5. sqlite.Start()
6. sqlite.Pipe <- []int64{}
   or sqlite.Send([]int64{})
7. sqlite.Stop()
8. sqlite.Close()
*/
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type SqliteData []int64
type SqliteHandle struct {
	Connection *sql.DB
	DBName     string
	IdTag      int64
	IdFirmware int
	Pipe       chan []int64
	Quit       chan bool
	WaitQuit   sync.WaitGroup
	NumTask    int
	TimeStamp  time.Time
	Error      error
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
		Pipe:    make(chan []int64, LENGTH_QUEUE),
		Quit:    make(chan bool),
		NumTask: runtime.NumCPU(),
	}
	return sqlite
}

func (s *SqliteHandle) Connect() (err error) {

	// set the database name according to patientId
	// fix database is locked
	// https://github.com/mattn/go-sqlite3/issues/148
	// _busy_timeout=XXX (default 5000 msec)
	// mode=rwc
	// cache=shared
	// Due to IO speed, _busy_timeout must be properly set.
	// Tested _busy_timeout
	// ---------------------------------------
	// | Target               | _busy_timour |
	// ---------------------------------------
	// | ArchLinux VirtualBox | 10000        |
	// | Window 7 64bit       | 5000         |
	// ---------------------------------------
	s.DBName = fmt.Sprintf("file:%s.db?mode=rwc&_busy_timeout=5000", s.PatientId)

	s.Connection, err = sql.Open("sqlite3", s.DBName)
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
	// Disable autocheckout
	if _, err := s.Connection.Exec(`PRAGMA wal_autocheckpoint=1000`); err != nil {
		return err
	}
	// For optmize use PRAGMA synchronous=NORMAL
	// default synchronous=FULL
	if _, err := s.Connection.Exec(`PRAGMA synchronous=NORMAL`); err != nil {
		return err
	}
	/* Create table tag
		TAG TABLE
	   ----------------------------------------------------------------------------------------------------------------------------------
	   | ID      | MNT   | UNIT | RESOLUTION | REFMAX | REFMIN | SamplingRate | ACTIVE  | Descriptor                      | Num Channel |
	   ----------------------------------------------------------------------------------------------------------------------------------
	   | PRI_KEY | TEXT  | TEXT | INTEGER    |  REAL  | REAL   | INTEGER(nsec)| INTEGER | TEXT                            | INTEGER     |
	   ----------------------------------------------------------------------------------------------------------------------------------
	   e.g.
	   ---------------------------------------------------------------------------------------------------------------------------------
	   | 1       | ECG1  |  mV  | 2048       |  1     |  -1    | 1000         | 0       | {"1": "LEAD_I", "2": "LEAD_II"} | 2          |
	   ---------------------------------------------------------------------------------------------------------------------------------
	   | 2       | ECG2  |   V  | 1024       |  5     |   0    | 2000         | 1       | {"1": "LEAD_VI", "4": "LEAD_II"}| 2          |
	   ---------------------------------------------------------------------------------------------------------------------------------
	   | 3       | SPO2  |  %   | 1024       |  0     |  100   | 3000         | 0       | {"5": "LEAD_I", "10": "LEAD_II"}| 2          |
	   ---------------------------------------------------------------------------------------------------------------------------------
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
	num_channel INTEGER,
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
	num_channel := len(desc)
	if num_channel == 0 {
		fmt.Errorf("Empty DescriptorType")
	}
	// convert to json
	jsonDdesc, _ := json.Marshal(desc)

	queryMeasurementStmt := `SELECT id FROM tag WHERE 
	mnt = ? AND unit = ? AND resolution = ? AND ref_min = ? AND 
	ref_max = ? AND sampling_rate = ? AND descriptor= ? AND num_channel = ?;`

	insertMeasurementStmt := `INSERT INTO tag 
	(mnt, unit, resolution, ref_min, ref_max, sampling_rate, descriptor,
	active, num_channel) VALUES (?,?,?,?,?,?,?,?,?);`

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
		string(jsonDdesc),
		num_channel).Scan(&s.IdTag)
	switch {
	case err == sql.ErrNoRows:
		// Insert new measurement in table tag and enable it
		p, err := s.Connection.Prepare(insertMeasurementStmt)
		if err != nil {
			return err
		}
		result, err := p.Exec(s.Measurement, s.Unit, s.Resolution, s.ReferenceMin, s.ReferenceMax, s.SamplingRate, string(jsonDdesc), 1, num_channel)
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
	 *   ---------------------------------------------------------------------
	 *   | TIME (nsec)  | SYNC    | CHANNEL_1 | CHANNEL_2 | .... | CHANNEL_N |
	 *   ---------------------------------------------------------------------
	 *   | PRI_KEY      | INTEGER | INTEGER   | INTEGER   | .... |  INTEGER  |
	 *   ---------------------------------------------------------------------
	 *   e.g.
	 *   ----------------------------------------------------------------------
	 *   | 109880980980 | 0       | 10908     | 1         | ...... | ..       |
	 *   ----------------------------------------------------------------------
	 *   | 1988-09-0909 | 0       | 78909     | 1         | ...... | ..       |
	 *   ----------------------------------------------------------------------
	 */
	measurementTableStmt := `CREATE TABLE IF NOT EXISTS %s_%d (
	time INTEGER NOT NULL,
	sync INTEGER DEFAULT 0,`

	for i, d := range desc {
		if i < num_channel-1 {
			measurementTableStmt += strings.ToLower(d) + " INTEGER NOT NULL,"
		} else {
			measurementTableStmt += strings.ToLower(d) + " INTEGER NOT NULL);"
		}

	}
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

	// wait-to-exit routine
	go func() {
		s.WaitQuit.Wait()
		s.Quit <- true
	}()

	// create SQL insertion
	var insertStmt string
	switch s.IdFirmware {
	case 1:
		insertStmt = fmt.Sprintf(
			`INSERT INTO %s_%d VALUES (?,0,?,?);`, s.Measurement, s.IdTag)
	case 2:
		insertStmt = fmt.Sprintf(
			`INSERT INTO %s_%d VALUES (?,0,?,?);`, s.Measurement, s.IdTag)
	case 3:
		insertStmt = fmt.Sprintf(
			`INSERT INTO %s_%d VALUES (?,0,?,?,?,?,?,?,?,?,?,?,?,?);`,
			s.Measurement, s.IdTag)
	default:
		s.Error = fmt.Errorf("No Match Firware Id")
		return
	}

	// init worker
	for i := 0; i < s.NumTask; i++ {
		s.WaitQuit.Add(1)
		// main routine
		go func(id int) {
			defer s.WaitQuit.Done()

			// create local sqlite connection
			conn, err := sql.Open("sqlite3", s.DBName)
			if err != nil {
				fmt.Println("Err sql.Open(): ", err)
				s.Error = fmt.Errorf("Err sql.Open(): ", err)
				return
			}
			defer conn.Close()

			// cache prepare statement for speed up
			var stmt *sql.Stmt
			stmt, err = conn.Prepare(insertStmt)
			if err != nil {
				fmt.Println("Err TX Prepare(): ", err)
				s.Error = fmt.Errorf("Err TX Prepare(): ", err)
				return
			}
			// init counter for transaction commit
			// init local variables
			counter := 0
			isBegin := false
			var timestamp int64

			defer func() {
				if isBegin {
					if _, err = conn.Exec(`COMMIT;`); err != nil {
						fmt.Println("Err TX Commit: ", err)
						s.Error = fmt.Errorf("Err TX Commit: ", err)
					}
					//fmt.Println("END routine ", id)
				}
			}()
			// main loop
			for data := range s.Pipe {
				// init transaction for counter = 0
				if isBegin == false {
					_, err = conn.Exec(`BEGIN;`)
					if err != nil {
						fmt.Println("Err TX Begin(): ", err)
						s.Error = fmt.Errorf("Err TX Begin(): ", err)
					}
					isBegin = true
				}
				timestamp = data[0]
				data = data[1:]

				// insert data
				switch s.IdFirmware {
				case 1:
					if len(data) >= 2 {
						if _, err = stmt.Exec(timestamp, data[0], data[1]); err != nil {
							fmt.Println("Err TX Exec: ", err)
							s.Error = fmt.Errorf("Err TX Exec: ", err)
						}
					}
				case 2:
					if len(data) >= 2 {
						if _, err = stmt.Exec(timestamp, data[0], data[1]); err != nil {
							fmt.Println("Err TX Exec: ", err)
							s.Error = fmt.Errorf("Err TX Exec: ", err)
						}
					}
				case 3:
					if len(data) >= 12 {
						if _, err = stmt.Exec(timestamp, data[0], data[1],
							data[2], data[3], data[4], data[5], data[6], data[7],
							data[8], data[9], data[10], data[11]); err != nil {
							fmt.Println("Err TX Exec: ", err)
							s.Error = fmt.Errorf("Err TX Exec: ", err)
						}
					}
				}
				counter++

				// periodically commit when every 1000 record
				if counter >= 1000 && isBegin == true {

					if _, err = conn.Exec(`COMMIT;`); err != nil {
						fmt.Println("Err TX Commit: ", err)
						s.Error = fmt.Errorf("Err TX Commit: ", err)
					}
					counter = 0
					isBegin = false

				}
			}

		}(i)
	}
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
	/*
		if _, err := s.Connection.Exec(`PRAGMA wal_checkpoint(PASSIVE);`); err != nil {
			fmt.Println("Err wal_checkpoint ", err)
		}*/
	s.Connection.Close()
}
