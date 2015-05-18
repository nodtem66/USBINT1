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
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
	. "github.com/nodtem66/usbint1/config"
	. "github.com/nodtem66/usbint1/webapi"
)

// Define Build LDFLAGS variable
var Version string
var Commit string

// global variable
var conf TomlConfig

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// new handle
	handle := New()
	if len(Version) > 0 {
		handle.Version = Version
	}
	if len(Commit) > 0 {
		handle.Commit = Commit
	}

	// load the toml configuration
	var conf TomlConfig
	if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
		log.Fatal(err)
	}
	// save to handle
	handle.Conf = &conf

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
	router := NewAPIRouter(handle)

	// start server
	fmt.Printf("start server %s [see %s]\n", conf.Server.Address, conf.Log.FileName)
	log.Fatal(http.ListenAndServe(conf.Server.Address, router))
}
