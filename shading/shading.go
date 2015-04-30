package shading

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"path/filepath"
	"strings"
	"time"
)

const SQLITE_MODE string = "file:%s?mode=rw"

type Handler struct {
	Root        string
	MinimumSync int
	Interval    time.Duration
	done        chan struct{}
}

// new handler from root path of database file
func NewHandler(root string) *Handler {
	handler := &Handler{
		Root:        root,
		Interval:    time.Minute,
		MinimumSync: 1,
		done:        make(chan struct{}),
	}
	return handler
}

// start operate with a fix interval
func (h *Handler) Start() {

	ticker := time.Tick(h.Interval)
	log.Printf("[Start Shading at %s every %s]", h.Root, h.Interval)
	go func() {
		for {
			select {
			case <-h.done:
				return
			case <-ticker:
				h.Operate()
			}
		}
	}()
}

// stop routine
func (h *Handler) Stop() {
	close(h.done)
}

// perform compact every table on every database file in root path
// note: every database file must be ended with `.db`
func (h *Handler) Operate() {
	files, _ := filepath.Glob(filepath.Join(h.Root, "*.db"))

	// for all patient database files
file_loop:
	for _, file := range files {
		var rows *sql.Rows
		var sqlite *sql.DB
		var err error
		patientId := filepath.Base(file)
		patientId = patientId[0 : len(patientId)-3]
		log.Printf("[Checking database %s]", patientId)
		// open database file with SQLITE_MODE
		if sqlite, err = sql.Open("sqlite3", fmt.Sprintf(SQLITE_MODE, file)); err != nil {
			log.Println(err)
		}
		defer sqlite.Close()

		tables := make([]string, 0)
		// query all table names from file
		rows, err = sqlite.Query(`SELECT name FROM sqlite_master WHERE type='table';`)
		defer rows.Close()
		for rows.Next() {
			var n string
			if err = rows.Scan(&n); err != nil {
				log.Println(err)
				continue file_loop
			}
			if strings.ToLower(n) != "tag" {
				tables = append(tables, n)
			}
		}
		if err = rows.Err(); err != nil {
			log.Println(err)
			continue file_loop
		}

		// for each table name delete synchronized record (sync > 0)
		for _, table := range tables {
			var result sql.Result
			if result, err = sqlite.Exec(
				fmt.Sprintf(`DELETE FROM %s WHERE sync >= %d;`, table, h.MinimumSync)); err != nil {
				log.Println(err)
				continue
			}
			// print number of deleted records
			if count, err := result.RowsAffected(); err == nil {
				log.Printf("[Compact %s clean %d record(s)]", table, count)
			}
			// force to vacuum database
			if _, err = sqlite.Exec("VACUUM;"); err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
