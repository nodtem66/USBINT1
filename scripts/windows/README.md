# Installation script
The usbint1 use [nssm](http://nssm.cc/) for service management
to use the scripts, select the folder depended on OS architecture:
* for Window 32bit use `./x86`
* for Window 64bit use `./x64`

## Install/Remove `usbint1` Service
* copy `usbint1.exe` to working path (e.g. D:\usbint1\)
  note that this path will be database path for web application
* edit the `env.txt` to working path  (e.g. D:\usbint1\)
* run install.bat` for install service
* run `uninstall.bat` for remove service

## Path installation
Install the path of script to environment

* run cmd.exe and type
  `C:\Windows\System32\rundll32.exe sysdm.cpl,EditEnvironmentVariables`
* edit `Path` variable and append current script path to the end (use semicolon for token)
  ...`;D:\usbint1\scripts\x64\` 

## Service Management
* use `start.bat` for start service
* use `stop.bat` for stop service
* use `config.bat` for edit service