# USBINT1
USB firmware in #golang for ARM6 (Raspberry Pi)

## Directory
* `cmd`:
  command line program to run daemon
* `config`:
  the set of config file and its reader
* `test`: 
  The collection of simple program to prove the proper protocol instruction.
* `firmware`:
  the set of firmwares and loader. Firmware is read the data from USB and send to `wrapper`
* `wrapper`:
  the set of wrapper and loader. Wrapper is wrap the streaming data from the firmware and send to `db` 
* `db`:
  the set of databases and loader. the database is endpoint to finish the steaming data

## Test (interrupt with libusb implemented with C)
* open test/usbint1.workspace with [Codelite](http://codelite.org/) 
* build and run

## Related projects
* [USBINT1-C8051F380](https://github.com/nodtem66/USBINT1-C8051F380): An USB firmware for Silab C8051F380
