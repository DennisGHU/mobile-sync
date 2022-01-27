@echo off
title Compiler
:restart
echo.
echo Compiling...
@echo on
go build -o server.exe && server.exe

pause
cls
goto :restart