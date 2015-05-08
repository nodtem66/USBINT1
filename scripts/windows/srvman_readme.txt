# Windows Service Manager (SrvMan)
Windows Service Manager is a small tool that simplifies all common tasks related to Windows services. It can create services (both Win32 and Legacy Driver) without restarting Windows, delete existing services and change service configuration. It has both GUI and Command-line modes. It can also be used to run arbitrary Win32 applications as services (when such service is stopped, main application window is closed automatically).

## SrvMan - Command Line Options
You can use SrvMan's Command Line interface to perform the following tasks:
* Create services
* Deleting services
* Start/stop/restart services
* Install & start a legacy driver with a single call

Note that when you run SrvMan with command-line arguments from non-console application (for example, from a Run dialog box), it displays the "Press any key to continue..." message and pauses just before exiting. This does not happen, when SrvMan is run from a console application (such as cmd.exe). To override this behavior, use the /pause:no switch.

## Creating services
Use the following command line to create services using SrvMan (parameters in brackets are optional):
```
srvman.exe add <file.exe/file.sys> [service name] [display name] [/type:<service type>] [/start:<start mode>] [/interactive:no] [/overwrite:yes]
Service name is an internal name used by Windows to reference the service. Display name is the name displayed in Windows Services snap-in. By default, both names are generated from the .exe or .sys file name, however, you can override it by specifying names explicitly.
```
**Service type can be one of the following:**

*`drv` - Create a kernel driver (selected by default for .sys files)
*`exe` - Create a Win32 service (selected by default for .exe files)
*`sharedexe` - Create a Win32 service with shared executable file
*`fsd` - Create a file system driver service
*`app` - Create a service running ordinary windows application (such as taskmgr.exe)

**Start mode is one of the following:**

*`boot` - The service is started by OS loader
*`sys` - The service is started by IoInitSystem() call
*`auto` - The service is started by Service Control Manager during startup
*`man` - The service is started manually (net start/net stop)
*`dis` - The service cannot be started
Win32 services are created as interactive by default. To create a non-interactive service, you should specify the /interactive:no parameter. Normally, if a specified service already exists, SrvMan reports an error and stops. However, if you specify the /overwrite:yes parameter, an existing service will be overwritten instead.

## Deleting services
Deleting services using SrvMan command line is quite obvious:
```
srvman.exe delete <service name>
Note that you need to specify the internal service name (same as used for net start command), not the display name.
```
## Starting/stopping/restarting services
You can control all types of services using SrvMan command line:
```
srvman.exe start <service name> [/nowait] [/delay:<delay in msec>]
srvman.exe stop <service name> [/nowait] [/delay:<delay in msec>]
srvman.exe restart <service name> [/delay:<delay in msec>]
```
Normally, SrvMan waits for the service to start. However, if you specify the /nowait parameter, SrvMan will return control immediately after the start/stop request was issued. Note that if you need SrvMan to wait before starting/stopping the service (for example, to switch to real-time log viewer window), you can use the /delay:<delay in msec> parameter.

## Testing legacy drivers
You can easily test your legacy driver by using the following command line:
```
srvman.exe run <driver.sys> [service name] [/copy:yes] [/overwrite:no] [/stopafter:<msec>]
This command creates (or overwrites) a service for a given legacy driver file and starts it. If you have specified the /copy:yes switch, the driver file will be copied to system32\drivers directory. If /overwrite:no is specified, DbgMan will return an error if the service (or the driver file in system32\drivers) already exists. If /after:<msec> is specified, the driver will be stopped msec milliseconds after successful start. You can use this switch to test driver load/unload cycle.
```
## Downloading
The latest version of SrvMan for both x86 and x64 systems can be downloaded here: srvman-1.0.zip. Source code can be downloaded here: srvman-1.0-src.zip. Note that you need BazisLib library to build the sources.

## Support
Feel free to ask questions about SrvMan or any other SysProgs tools on the SysProgs.org forum.