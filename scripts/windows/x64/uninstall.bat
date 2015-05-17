@echo off
FOR /L %%i IN (1,1,10) DO (
nssm.exe remove usbint1_%%i confirm
)
pause