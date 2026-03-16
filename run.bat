echo off

rem Launches RelayServer with go run
rem Could be better (like exiting after running)

where /q go
if ERRORLEVEL 1 (
	echo go could not be found, please ensure it is installed and exported to PATH
	exit 1
)

cd main
go mod tidy
go run .

