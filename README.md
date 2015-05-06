# USBINT1 [![Build Status](https://travis-ci.org/nodtem66/USBINT1.svg)](https://travis-ci.org/nodtem66/USBINT1) [![GoDoc](http://godoc.org/github.com/nodtem66/USBINT1?status.png)](http://godoc.org/github.com/nodtem66/USBINT1)
USB Host Firmware in #golang for Silab C8051F380 [USBINT1-C8051F380](https://github.com/nodtem66/USBINT1-C8051F380)

## Test OS
* ARM6 (Raspberry Pi)
* Window7 64bit 

## Requirements
* **libusb-1.0**
  For Windows, [Zadig](http://zadig.akeo.ie/) is recommended for install libusb-based driver
  
## Directory
* `cmd`:
  command line programs to run daemon
* `test`: 
  The collection of simple program to prove the proper protocol instruction.
* `firmware`:
  the set of firmwares and loader.
* `db`:
  the set of databases and loader. the database is endpoint to finish the steaming data
  the current db engine is sqlite3
* `shading`:
  the set of shading package to minimize size of database file after synchronization
* `sync`:
  worker for sqlite3-mariadb synchronization
* `config`:
  an model of configuration file used globally in `usbapi`, `usbsync`, and `usbshad`

## Commands in `cmd`
* `usbint`:
   command line program to run daemon with target vid, pid, and patient id
* `usbtest`:
   command line to test `kylelemons/gousb`
* `usbapi`:
   command line to run webserver, serving the sqlite3 db with RESTful api
* `usbsync`:
   command line to run sync worker; move from device database file (sqlite3) to endpoint server (MariaDB in this case)
* `usbshad`:
   command line to run shaing work; minimize to database file by deleting and compacting sqlite3 file

## Bugs
*  Exception 0xC0000005 on Window7 64bit
   see go-sqlite3 tickets: [mattn/go-sqlite3#163](https://github.com/mattn/go-sqlite3/issues/163) [golang/go#9356](https://github.com/golang/go/issues/9356) 

## Test (interrupt with libusb implemented with C)
* open test/usbint1.workspace with [Codelite](http://codelite.org/) 
* build and run

## Related projects
* [USBINT1-C8051F380](https://github.com/nodtem66/USBINT1-C8051F380): An USB firmware for Silab C8051F380