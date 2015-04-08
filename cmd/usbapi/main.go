/* command line for read the sqlite db file
 * The following API is defined:
 * 1. /patient
 *      List of all patient db files in devices
 * 2. /patient/:patientId/tag
 *      List of all tags in PatientId
 * 3. /patient/:patientId/tag?query
 *      List of all tags in PatientId with condition
 * 4. /patient/:patientId/tag/:tagId
 *      List of all descriptions in tagId
 * 5. /patient/:patientId/mnt
 *      List of all measurement units in patientId
 * 6. /patient/:patientId/mnt/:mntId
 *      Description of measurement unit
 * 7. /patient/:patientId/mnt/:mntId?query
 *      List of signal from measurement unit with condition
 */
package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Define Build LDFLAGS variable
var Version string
var Commit string

// Configuration Type
type TomlConfig struct {
	Device device
	DB     database `toml:"database"`
	Server server
	Log    loginfo `toml:"log"`
}

type device struct {
	Name string `toml:"name"`
	Org  string `toml:"organization"`
	Desc string `toml:"description"`
}
type database struct {
	Path string `toml:"path"`
}

type server struct {
	Address string `toml:"address"`
}
type loginfo struct {
	FileName string `toml:"file"`
}

// global variable
var conf TomlConfig

// Index page print program name and version
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "USBAPI [Version: %s] [Commit: %s]\n", Version, Commit)
}

// Handler for /patient
// List of all patient db files in devices
func GetPatients(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	files, err := ioutil.ReadDir(conf.DB.Path)
	if err != nil {
		log.println(err)
	}
	fmt.Fprintf(w, "%#v", files)
}

// Handler for
// 1. /patient/:patientId/tag
//   List of all tags in PatientId
// 2. /patient/:patientId/tag?query
//   List of all tags in PatientId with condition
func GetTags(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintln(w, ps)
	fmt.Fprintln(w, r.URL.Query())
}

// Handler for /patient/:patientId/tag/:tagId
// List of all descriptions in tagId
func GetTag(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintln(w, ps)
}

// Handler for /patient/:patientId/mnt
// List of all measurement units in patientId
func GetMeasurements(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintln(w, ps)
}

// Handler for
// 1. /patient/:patientId/mnt/:mntId
//   Description of measurement unit
// 2. /patient/:patientId/mnt/:mntId?query
//   List of signal from measurement unit with condition
func GetMeasurement(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintln(w, ps)
	fmt.Fprintln(w, r.URL.Query())
}

func main() {
	// load the toml configuration
	if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
		log.Fatal(err)
	}

	// redirect log to file
	if len(conf.Log.FileName) != 0 {
		logfile, err := os.OpenFile(conf.Log.FileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer logfile.Close()

		// setting file to log
		log.SetOutput(logfile)
	}
	log.SetPrefix("[USB_API] ")
	log.Printf("[Start] [database = %s] [server = %s] [device = %s]",
		conf.DB.Path, conf.Server.Address, conf.Device.Name)
	// config the router
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/patient", GetPatients)
	router.GET("/patient/:patientId/tag", GetTags)
	router.GET("/patient/:patientId/tag/:tagId", GetTag)
	router.GET("/patient/:patientId/mnt", GetMeasurements)
	router.GET("/patient/:patientId/mnt/:mntId", GetMeasurement)

	// start server
	log.Fatal(http.ListenAndServe(conf.Server.Address, router))
}
