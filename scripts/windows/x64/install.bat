@echo off
SET /P DIR=<env.txt
nssm.exe install usbint1 %DIR%\usbint1.exe
nssm.exe set usbint1 AppParameters T001
nssm.exe set usbint1 AppStdout %DIR%\usbint1.log
nssm.exe set usbint1 AppStderr %DIR%\usbint1.log
nssm.exe set usbint1 Start SERVICE_DEMAND_START
nssm.exe set usbint1 AppStdoutCreationDisposition 4
nssm.exe set usbint1 AppStderrCreationDisposition 4
nssm.exe set usbint1 AppRotateFiles 1
nssm.exe set usbint1 AppRotateOnline 0
nssm.exe set usbint1 AppRotateSeconds 86400
nssm.exe set usbint1 AppRotateBytes 1048576
echo.
pause