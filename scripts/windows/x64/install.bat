@echo off
SET /P DIR=<env.txt
FOR /L %%i IN (1,1,10) DO (
nssm.exe install usbint1_%%i %DIR%\usbint1.exe
nssm.exe set usbint1_%%i AppParameters T001
nssm.exe set usbint1_%%i AppStdout %DIR%\usbint1.log
nssm.exe set usbint1_%%i AppStderr %DIR%\usbint1.log
nssm.exe set usbint1_%%i Start SERVICE_DEMAND_START
nssm.exe set usbint1_%%i AppStdoutCreationDisposition 4
nssm.exe set usbint1_%%i AppStderrCreationDisposition 4
nssm.exe set usbint1_%%i AppRotateFiles 1
nssm.exe set usbint1_%%i AppRotateOnline 0
nssm.exe set usbint1_%%i AppRotateSeconds 86400
nssm.exe set usbint1_%%i AppRotateBytes 1048576
echo.
)
pause